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

type StudentRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{pool: pool}
}

func (r *StudentRepository) FindAll(ctx context.Context, filter model.StudentFilterQuery) ([]model.StudentItemResponse, int, error) {
	var conditions []string
	var args []any
	argIndex := 1

	// Hanya mahasiswa yang belum di-soft-delete
	conditions = append(conditions, "deleted_at IS NULL")

	if filter.Prodi != "" {
		conditions = append(conditions, fmt.Sprintf("prodi ILIKE $%d", argIndex))
		args = append(args, "%"+filter.Prodi+"%")
		argIndex++
	}

	if filter.Angkatan > 0 {
		conditions = append(conditions, fmt.Sprintf("angkatan = $%d", argIndex))
		args = append(args, filter.Angkatan)
		argIndex++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(nim ILIKE $%d OR nama ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Hitung total data
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM students WHERE %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Tentukan sorting
	orderBy := "nama ASC"
	switch filter.Sort {
	case "nama":
		orderBy = "nama ASC"
	case "-nama":
		orderBy = "nama DESC"
	case "ipk_terakhir":
		orderBy = "ipk_terakhir ASC"
	case "-ipk_terakhir":
		orderBy = "ipk_terakhir DESC"
	default:
		orderBy = "id ASC"
	}

	limit := filter.PerPage
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	query := fmt.Sprintf(
		"SELECT id, nim, nama, prodi, angkatan, ipk_terakhir FROM students WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d",
		whereClause, orderBy, argIndex, argIndex+1,
	)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	students := []model.StudentItemResponse{}
	for rows.Next() {
		var s model.StudentItemResponse
		if err := rows.Scan(&s.ID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir); err != nil {
			return nil, 0, err
		}
		students = append(students, s)
	}

	return students, total, nil
}

func (r *StudentRepository) FindByID(ctx context.Context, id int) (*model.Student, error) {
	query := `
		SELECT s.id, s.user_id, s.nim, s.nama, u.email, s.prodi, s.angkatan, s.ipk_terakhir, s.deleted_at
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = $1 AND s.deleted_at IS NULL
	`
	var s model.Student
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Email, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepository) FindByUserID(ctx context.Context, userID int) (*model.Student, error) {
	query := `
		SELECT s.id, s.user_id, s.nim, s.nama, u.email, s.prodi, s.angkatan, s.ipk_terakhir, s.deleted_at
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE s.user_id = $1
	`
	var s model.Student
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Email, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepository) FindByNIM(ctx context.Context, nim string) (*model.Student, error) {
	query := `
		SELECT s.id, s.user_id, s.nim, s.nama, u.email, s.prodi, s.angkatan, s.ipk_terakhir, s.deleted_at
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE s.nim = $1
	`
	var s model.Student
	err := r.pool.QueryRow(ctx, query, nim).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Email, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepository) ExistsByNIM(ctx context.Context, nim string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM students WHERE nim = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, nim).Scan(&exists)
	return exists, err
}

func (r *StudentRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *StudentRepository) Create(ctx context.Context, tx pgx.Tx, s *model.Student) (int, error) {
	query := `
		INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	var id int
	err := tx.QueryRow(ctx, query, s.UserID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir).Scan(&id)
	return id, err
}

func (r *StudentRepository) Update(ctx context.Context, id int, req model.UpdateStudentRequest) error {
	var ipk float64
	if req.IPKTerakhir != nil {
		ipk = *req.IPKTerakhir
	}

	query := `
		UPDATE students
		SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4
		WHERE id = $5 AND deleted_at IS NULL
	`
	cmdTag, err := r.pool.Exec(ctx, query, req.Nama, req.Prodi, req.Angkatan, ipk, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("student not found or soft-deleted")
	}
	return nil
}

func (r *StudentRepository) SoftDelete(ctx context.Context, id int) error {
	query := `UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("student not found or already deleted")
	}
	return nil
}

func (r *StudentRepository) GetStudentCourses(ctx context.Context, studentID int) ([]model.CourseDetail, error) {
	query := `
		SELECT e.id, c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, e.tahun_akademik, e.created_at
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.student_id = $1
		ORDER BY e.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := []model.CourseDetail{}
	for rows.Next() {
		var cd model.CourseDetail
		if err := rows.Scan(&cd.EnrollmentID, &cd.ID, &cd.KodeMK, &cd.NamaMK, &cd.SKS, &cd.Semester, &cd.TahunAkademik, &cd.CreatedAt); err != nil {
			return nil, err
		}
		courses = append(courses, cd)
	}

	return courses, nil
}
