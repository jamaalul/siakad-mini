package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"siakad-mini/app/model"
)

var (
	ErrProjectLimit    = errors.New("jumlah proyek sudah mencapai batas")
	ErrProjectNotFound = errors.New("proyek tidak ditemukan")
)

// MaxProjectsPerPortfolio membatasi pertumbuhan array projects menjadi maksimal 20 item.
const MaxProjectsPerPortfolio = 20

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	FindByStudent(ctx context.Context, studentID int) (model.Portfolio, error)
	Upsert(ctx context.Context, studentID int, req model.PutPortfolioRequest) (model.Portfolio, error)
	AddProject(ctx context.Context, studentID int, proj model.Project) (model.Portfolio, error)
	DeleteProject(ctx context.Context, studentID int, projectID string) error
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
	URL          string        `bson:"url"`
	Year         int           `bson:"year"`
}

type certificateDocument struct {
	ID       bson.ObjectID `bson:"_id"`
	Name     string        `bson:"name"`
	Issuer   string        `bson:"issuer"`
	IssuedAt time.Time     `bson:"issued_at"`
}

func (d portfolioDocument) toModel() model.Portfolio {
	projects := make([]model.Project, 0, len(d.Projects))
	for _, p := range d.Projects {
		techs := p.Technologies
		if techs == nil {
			techs = []string{}
		}
		projects = append(projects, model.Project{
			ID: p.ID.Hex(), Title: p.Title, Technologies: techs,
			URL: p.URL, Year: p.Year,
		})
	}
	certs := make([]model.Certificate, 0, len(d.Certificates))
	for _, c := range d.Certificates {
		certs = append(certs, model.Certificate{
			ID: c.ID.Hex(), Name: c.Name, Issuer: c.Issuer, IssuedAt: c.IssuedAt,
		})
	}
	return model.Portfolio{
		ID: d.ID.Hex(), StudentID: d.StudentID, Headline: d.Headline, Bio: d.Bio,
		Projects: projects, Certificates: certs,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

type portfolioMongoRepository struct {
	coll *mongo.Collection
}

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{coll: db.Collection("student_portfolios")}
}

// EnsureIndexes membuat index unik untuk student_id dan multikey index untuk projects.technologies.
func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// Satu mahasiswa hanya boleh memiliki SATU dokumen portofolio.
			Keys:    bson.D{{Key: "student_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("student_portfolio_unique"),
		},
		{
			// Index multikey untuk mempercepat filter dan agregasi statistik teknologi.
			Keys:    bson.D{{Key: "projects.technologies", Value: 1}},
			Options: options.Index().SetName("portfolio_tech_multikey"),
		},
	})
	if err != nil {
		return fmt.Errorf("membuat index student_portfolios: %w", err)
	}
	return nil
}

// FindByStudent mengambil dokumen portofolio mahasiswa berdasarkan student_id.
func (r *portfolioMongoRepository) FindByStudent(ctx context.Context, studentID int) (model.Portfolio, error) {
	var doc portfolioDocument
	if err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrNotFound
		}
		return model.Portfolio{}, fmt.Errorf("mengambil portofolio: %w", err)
	}
	return doc.toModel(), nil
}

// Upsert membuat atau memperbarui profil portofolio mahasiswa.
// Digunakan $set untuk field yang diperbarui dan $setOnInsert untuk field inisialisasi awal.
func (r *portfolioMongoRepository) Upsert(ctx context.Context, studentID int, req model.PutPortfolioRequest) (model.Portfolio, error) {
	t := now()
	update := bson.M{
		"$set": bson.M{
			"headline":   req.Headline,
			"bio":        req.Bio,
			"updated_at": t,
		},
		"$setOnInsert": bson.M{
			"student_id":   studentID,
			"projects":     []projectDocument{},
			"certificates": []certificateDocument{},
			"created_at":   t,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var doc portfolioDocument
	if err := r.coll.FindOneAndUpdate(ctx, bson.M{"student_id": studentID}, update, opts).Decode(&doc); err != nil {
		return model.Portfolio{}, fmt.Errorf("menyimpan portofolio: %w", err)
	}
	return doc.toModel(), nil
}

// AddProject menambahkan satu proyek ke array projects secara atomik.
// Filter "projects.19" memastikan array belum berisi 20 elemen sebelum push dilakukan.
func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, proj model.Project) (model.Portfolio, error) {
	doc := projectDocument{
		ID:           bson.NewObjectID(),
		Title:        proj.Title,
		Technologies: proj.Technologies,
		URL:          proj.URL,
		Year:         proj.Year,
	}

	limitField := fmt.Sprintf("projects.%d", MaxProjectsPerPortfolio-1)
	var updated portfolioDocument
	err := r.coll.FindOneAndUpdate(ctx,
		bson.M{"student_id": studentID, limitField: bson.M{"$exists": false}},
		bson.M{"$push": bson.M{"projects": doc}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err == nil {
		return updated.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
	}

	// Bedakan "portofolio tidak ada" vs "batas tercapai"
	count, cerr := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if cerr != nil {
		return model.Portfolio{}, fmt.Errorf("memeriksa portofolio: %w", cerr)
	}
	if count == 0 {
		return model.Portfolio{}, ErrNotFound
	}
	return model.Portfolio{}, ErrProjectLimit
}

// DeleteProject menghapus satu proyek dari array projects menggunakan $pull.
func (r *portfolioMongoRepository) DeleteProject(ctx context.Context, studentID int, projectID string) error {
	oid, err := parseID(projectID)
	if err != nil {
		return ErrInvalidID
	}
	res, err := r.coll.UpdateOne(ctx,
		bson.M{"student_id": studentID},
		bson.M{"$pull": bson.M{"projects": bson.M{"_id": oid}}},
	)
	if err != nil {
		return fmt.Errorf("menghapus proyek: %w", err)
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	if res.ModifiedCount == 0 {
		return ErrProjectNotFound
	}
	return nil
}
