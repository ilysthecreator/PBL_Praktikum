package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"UTS/app/model"
)

type CourseRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) *CourseRepository {
	return &CourseRepository{pool: pool}
}

func (r *CourseRepository) FindAll(ctx context.Context, filter model.CourseFilterQuery) ([]model.Course, error) {
	var whereConditions []string
	var args []any
	argIndex := 1

	if filter.Semester > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("c.semester = $%d", argIndex))
		args = append(args, filter.Semester)
		argIndex++
	}

	if filter.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	havingClause := ""
	if filter.Available {
		havingClause = "HAVING (c.kuota - COUNT(e.id)::int) > 0"
	}

	query := fmt.Sprintf(`
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		       COUNT(e.id)::int AS terisi,
		       (c.kuota - COUNT(e.id)::int) AS sisa_kuota
		FROM courses c
		LEFT JOIN enrollments e ON c.id = e.course_id
		%s
		GROUP BY c.id
		%s
		ORDER BY c.kode_mk ASC
	`, whereClause, havingClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := []model.Course{}
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.SisaKuota); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}

	return courses, nil
}

func (r *CourseRepository) FindByID(ctx context.Context, id int) (*model.Course, error) {
	query := `
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		       COUNT(e.id)::int AS terisi,
		       (c.kuota - COUNT(e.id)::int) AS sisa_kuota
		FROM courses c
		LEFT JOIN enrollments e ON c.id = e.course_id
		WHERE c.id = $1
		GROUP BY c.id
	`
	var c model.Course
	err := r.pool.QueryRow(ctx, query, id).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.SisaKuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// FindByIDForUpdate mengunci row data mata kuliah untuk transaksi pendaftaran KRS
func (r *CourseRepository) FindByIDForUpdate(ctx context.Context, tx pgx.Tx, id int) (*model.Course, error) {
	query := `SELECT id, kode_mk, nama_mk, sks, semester, kuota FROM courses WHERE id = $1 FOR UPDATE`
	var c model.Course
	err := tx.QueryRow(ctx, query, id).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *CourseRepository) GetCurrentEnrolledCountInTx(ctx context.Context, tx pgx.Tx, courseID int) (int, error) {
	query := `SELECT COUNT(*) FROM enrollments WHERE course_id = $1`
	var count int
	err := tx.QueryRow(ctx, query, courseID).Scan(&count)
	return count, err
}
