package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"api-students/app/model"
)

type NilaiRepository interface {
	FindByStudentID(ctx context.Context, studentID int) ([]model.Nilai, error)
	Create(ctx context.Context, n model.Nilai) (model.Nilai, error)
}

type nilaiPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewNilaiRepository(pool *pgxpool.Pool) NilaiRepository {
	return &nilaiPostgresRepository{pool: pool}
}

func (r *nilaiPostgresRepository) FindByStudentID(ctx context.Context, studentID int) ([]model.Nilai, error) {
	query := `
		SELECT id_nilai, nama_mata_kuliah, nilai, id_student
		FROM nilai
		WHERE id_student = $1
		ORDER BY id_nilai ASC
	`
	rows, err := r.pool.Query(ctx, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar nilai: %w", err)
	}
	defer rows.Close()

	var list []model.Nilai
	for rows.Next() {
		var n model.Nilai
		if err := rows.Scan(&n.IDNilai, &n.NamaMataKuliah, &n.Nilai, &n.IDStudent); err != nil {
			return nil, fmt.Errorf("membaca baris nilai: %w", err)
		}
		list = append(list, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi baris nilai: %w", err)
	}

	if list == nil {
		list = []model.Nilai{}
	}

	return list, nil
}

func (r *nilaiPostgresRepository) Create(ctx context.Context, n model.Nilai) (model.Nilai, error) {
	query := `
		INSERT INTO nilai (nama_mata_kuliah, nilai, id_student)
		VALUES ($1, $2, $3)
		RETURNING id_nilai
	`
	err := r.pool.QueryRow(ctx, query, n.NamaMataKuliah, n.Nilai, n.IDStudent).Scan(&n.IDNilai)
	if err != nil {
		return model.Nilai{}, fmt.Errorf("menyimpan nilai: %w", err)
	}

	return n, nil
}
