package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"latihan-fiber/app/model"
)

var ErrProjectLimit = errors.New("jumlah proyek sudah mencapai batas")

const (
	PortfolioCollection     = "student_portfolios"
	MaxProjectsPerPortfolio = 20
)

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	GetByStudentID(ctx context.Context, studentID int) (model.Portfolio, error)
	Upsert(ctx context.Context, studentID int, headline, bio string, certs []model.Certificate) (model.Portfolio, error)
	AddProject(ctx context.Context, studentID int, proj model.Project) (model.Portfolio, error)
	DeleteProject(ctx context.Context, studentID int, projectID string) error
	GetTopTechnologies(ctx context.Context) ([]model.TechnologyStat, error)
}

type portfolioDocument struct {
	ID           bson.ObjectID         `bson:"_id,omitempty"`
	StudentID    int                   `bson:"student_id"`
	Headline     string                `bson:"headline"`
	Bio          string                `bson:"bio"`
	Projects     []projectDocument     `bson:"projects"`
	Certificates []certificateDocument `bson:"certificates"`
	CreatedAt    time.Time             `bson:"created_at"`
	UpdatedAt    time.Time             `bson:"updated_at"`
}

type projectDocument struct {
	ID           bson.ObjectID `bson:"_id"`
	Title        string        `bson:"title"`
	Technologies []string      `bson:"technologies"`
	Link         string        `bson:"link,omitempty"`
	Year         int           `bson:"year"`
}

type certificateDocument struct {
	ID     bson.ObjectID `bson:"_id"`
	Name   string        `bson:"name"`
	Issuer string        `bson:"issuer"`
	Date   string        `bson:"date"`
}

func (d portfolioDocument) toModel() model.Portfolio {
	projects := make([]model.Project, 0, len(d.Projects))
	for _, p := range d.Projects {
		techs := p.Technologies
		if techs == nil {
			techs = []string{}
		}
		projects = append(projects, model.Project{
			ID:           p.ID.Hex(),
			Title:        p.Title,
			Technologies: techs,
			Link:         p.Link,
			Year:         p.Year,
		})
	}

	certs := make([]model.Certificate, 0, len(d.Certificates))
	for _, c := range d.Certificates {
		certs = append(certs, model.Certificate{
			ID:     c.ID.Hex(),
			Name:   c.Name,
			Issuer: c.Issuer,
			Date:   c.Date,
		})
	}

	return model.Portfolio{
		ID:           d.ID.Hex(),
		StudentID:    d.StudentID,
		Headline:     d.Headline,
		Bio:          d.Bio,
		Projects:     projects,
		Certificates: certs,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

type portfolioMongoRepository struct {
	coll *mongo.Collection
}

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{
		coll: db.Collection(PortfolioCollection),
	}
}

func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "student_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("student_unique"),
		},
		{
			Keys:    bson.D{{Key: "projects.technologies", Value: 1}},
			Options: options.Index().SetName("project_technologies"),
		},
	})
	if err != nil {
		return fmt.Errorf("membuat index portofolio: %w", err)
	}
	return nil
}

func (r *portfolioMongoRepository) GetByStudentID(ctx context.Context, studentID int) (model.Portfolio, error) {
	var doc portfolioDocument
	err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrNotFound
		}
		return model.Portfolio{}, fmt.Errorf("mencari portofolio: %w", err)
	}
	return doc.toModel(), nil
}

func (r *portfolioMongoRepository) Upsert(ctx context.Context, studentID int, headline, bio string, certs []model.Certificate) (model.Portfolio, error) {
	t := now()
	setFields := bson.M{
		"headline":   headline,
		"bio":        bio,
		"updated_at": t,
	}

	if certs != nil {
		certDocs := make([]certificateDocument, 0, len(certs))
		for _, c := range certs {
			oid := bson.NewObjectID()
			if c.ID != "" {
				if parsed, err := bson.ObjectIDFromHex(c.ID); err == nil {
					oid = parsed
				}
			}
			certDocs = append(certDocs, certificateDocument{
				ID:     oid,
				Name:   c.Name,
				Issuer: c.Issuer,
				Date:   c.Date,
			})
		}
		setFields["certificates"] = certDocs
	}

	setOnInsert := bson.M{
		"student_id": studentID,
		"created_at": t,
		"projects":   []projectDocument{},
	}
	if certs == nil {
		setOnInsert["certificates"] = []certificateDocument{}
	}

	update := bson.M{
		"$set":         setFields,
		"$setOnInsert": setOnInsert,
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var updated portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"student_id": studentID}, update, opts).Decode(&updated)
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("upsert portofolio: %w", err)
	}
	return updated.toModel(), nil
}

func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, proj model.Project) (model.Portfolio, error) {
	limitField := "projects." + strconv.Itoa(MaxProjectsPerPortfolio-1)

	doc := projectDocument{
		ID:           bson.NewObjectID(),
		Title:        proj.Title,
		Technologies: proj.Technologies,
		Link:         proj.Link,
		Year:         proj.Year,
	}

	filter := bson.M{
		"student_id": studentID,
		limitField:   bson.M{"$exists": false},
	}
	update := bson.M{
		"$push": bson.M{"projects": doc},
		"$set":  bson.M{"updated_at": now()},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err == nil {
		return updated.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
	}

	// Tidak cocok: bedakan apakah portofolio belum dibuat atau limit tercapai
	count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("memeriksa portofolio: %w", err)
	}
	if count == 0 {
		return model.Portfolio{}, ErrNotFound
	}
	return model.Portfolio{}, ErrProjectLimit
}

func (r *portfolioMongoRepository) DeleteProject(ctx context.Context, studentID int, projectID string) error {
	oid, err := parseID(projectID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"student_id":   studentID,
		"projects._id": oid,
	}
	update := bson.M{
		"$pull": bson.M{"projects": bson.M{"_id": oid}},
		"$set":  bson.M{"updated_at": now()},
	}

	res, err := r.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("menghapus proyek: %w", err)
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *portfolioMongoRepository) GetTopTechnologies(ctx context.Context) ([]model.TechnologyStat, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$projects"}},
		{{Key: "$unwind", Value: "$projects.technologies"}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$projects.technologies",
			"count": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "count", Value: -1},
			{Key: "_id", Value: 1},
		}}},
		{{Key: "$limit", Value: 10}},
		{{Key: "$project", Value: bson.M{
			"_id":        0,
			"technology": "$_id",
			"count":      1,
		}}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("agregasi teknologi: %w", err)
	}
	defer cursor.Close(ctx)

	var results []model.TechnologyStat
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("membaca hasil agregasi teknologi: %w", err)
	}
	if results == nil {
		results = []model.TechnologyStat{}
	}
	return results, nil
}
