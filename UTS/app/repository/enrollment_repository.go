package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"UTS/app/model"
)

type EnrollmentRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) *EnrollmentRepository {
	return &EnrollmentRepository{pool: pool}
}

func (r *EnrollmentRepository) IsEnrolledInTA(ctx context.Context, tx pgx.Tx, studentID, courseID int, tahunAkademik string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM enrollments 
			WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3
		)
	`
	var exists bool
	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, query, studentID, courseID, tahunAkademik).Scan(&exists)
	} else {
		err = r.pool.QueryRow(ctx, query, studentID, courseID, tahunAkademik).Scan(&exists)
	}
	return exists, err
}

func (r *EnrollmentRepository) GetTotalSKSInTA(ctx context.Context, tx pgx.Tx, studentID int, tahunAkademik string) (int, error) {
	query := `
		SELECT COALESCE(SUM(c.sks), 0)::int
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.student_id = $1 AND e.tahun_akademik = $2
	`
	var totalSKS int
	var err error
	if tx != nil {
		err = tx.QueryRow(ctx, query, studentID, tahunAkademik).Scan(&totalSKS)
	} else {
		err = r.pool.QueryRow(ctx, query, studentID, tahunAkademik).Scan(&totalSKS)
	}
	return totalSKS, err
}

func (r *EnrollmentRepository) Create(ctx context.Context, tx pgx.Tx, studentID, courseID int, tahunAkademik string) (*model.Enrollment, error) {
	query := `
		INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		VALUES ($1, $2, $3)
		RETURNING id, student_id, course_id, tahun_akademik, created_at
	`
	var e model.Enrollment
	err := tx.QueryRow(ctx, query, studentID, courseID, tahunAkademik).Scan(
		&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EnrollmentRepository) FindByID(ctx context.Context, id int) (*model.Enrollment, error) {
	query := `SELECT id, student_id, course_id, tahun_akademik, created_at FROM enrollments WHERE id = $1`
	var e model.Enrollment
	err := r.pool.QueryRow(ctx, query, id).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *EnrollmentRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM enrollments WHERE id = $1`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("enrollment tidak ditemukan")
	}
	return nil
}
