package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"

	"latihan-fiber/app/model"
	"latihan-fiber/config"
	"latihan-fiber/database"
)

func main() {
	outboxID := flag.Int64("outbox", 0, "ID outbox yang akan di-redeliver")
	copies := flag.Int("copies", 5, "Jumlah salinan task yang dikirim ke Redis")
	flag.Parse()

	if *outboxID <= 0 {
		log.Fatal("flag -outbox wajib diisi dengan ID outbox yang valid (> 0)")
	}

	config.LoadEnv()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("gagal koneksi ke database: %v", err)
	}
	defer pool.Close()

	var eventType string
	var payload []byte
	err = pool.QueryRow(ctx, "SELECT event_type, payload FROM outbox WHERE id = $1", *outboxID).
		Scan(&eventType, &payload)
	if err != nil {
		log.Fatalf("mengambil outbox id %d: %v", *outboxID, err)
	}

	job := model.NotificationJob{
		OutboxID:  *outboxID,
		EventType: eventType,
		Payload:   payload,
	}

	jobBytes, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("serialize job: %v", err)
	}

	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()

	for i := 1; i <= *copies; i++ {
		// Gunakan task ID unik per salinan agar tidak diblokir oleh ErrTaskIDConflict
		task := asynq.NewTask(model.TaskSendNotification, jobBytes,
			asynq.TaskID(fmt.Sprintf("redeliver-%d-%d-%d", *outboxID, i, time.Now().UnixNano())),
			asynq.Queue(model.QueueNotifications),
			asynq.MaxRetry(5),
			asynq.Timeout(30*time.Second),
			asynq.Retention(24*time.Hour),
		)

		info, err := client.EnqueueContext(ctx, task)
		if err != nil {
			log.Fatalf("gagal enqueue salinan ke-%d: %v", i, err)
		}
		fmt.Printf("[OK] Salinan %d/%d berhasil di-enqueue: ID=%s, Queue=%s\n", i, *copies, info.ID, info.Queue)
	}

	fmt.Printf("\nSelesai! Berhasil mengirim %d salinan task untuk outbox ID %d ke Redis.\n", *copies, *outboxID)
}
