package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"siakad-mini/app/model"
)

// ErrNotFound dan ErrDuplicate sudah ada di student_repository.go.
var (
	ErrInvalidID    = errors.New("format id tidak valid")
	ErrNotOwner     = errors.New("bukan pemilik data")
	ErrOwnReview    = errors.New("tidak boleh menilai ulasan sendiri")
	ErrAlreadyVoted = errors.New("sudah pernah menilai ulasan ini")
	ErrReplyLimit   = errors.New("jumlah balasan sudah mencapai batas")
)

// MaxRepliesPerReview membatasi pertumbuhan array replies. Satu dokumen
// MongoDB maksimal 16 MB; array tanpa batas suatu saat akan menabrak batas itu.
const MaxRepliesPerReview = 50

type ReviewRepository interface {
	EnsureIndexes(ctx context.Context) error
}

// reviewDocument adalah bentuk ulasan SEBAGAIMANA TERSIMPAN di MongoDB.
// Tag bson menentukan nama field di database.
type reviewDocument struct {
	ID           bson.ObjectID   `bson:"_id,omitempty"`
	CourseID     int             `bson:"course_id"`
	StudentID    int             `bson:"student_id"`
	StudentName  string          `bson:"student_name"`
	Rating       int             `bson:"rating"`
	Comment      string          `bson:"comment"`
	Tags         []string        `bson:"tags"`
	Replies      []replyDocument `bson:"replies"`
	HelpfulCount int             `bson:"helpful_count"`
	HelpfulBy    []int           `bson:"helpful_by"`
	CreatedAt    time.Time       `bson:"created_at"`
	UpdatedAt    time.Time       `bson:"updated_at"`
}

type replyDocument struct {
	ID          bson.ObjectID `bson:"_id"`
	StudentID   int           `bson:"student_id"`
	StudentName string        `bson:"student_name"`
	Body        string        `bson:"body"`
	CreatedAt   time.Time     `bson:"created_at"`
}

func (d reviewDocument) toModel() model.Review {
	replies := make([]model.Reply, 0, len(d.Replies))
	for _, r := range d.Replies {
		replies = append(replies, model.Reply{
			ID: r.ID.Hex(), StudentID: r.StudentID, StudentName: r.StudentName,
			Body: r.Body, CreatedAt: r.CreatedAt,
		})
	}
	tags := d.Tags
	if tags == nil {
		tags = []string{}
	}
	return model.Review{
		ID: d.ID.Hex(), CourseID: d.CourseID, StudentID: d.StudentID, StudentName: d.StudentName,
		Rating: d.Rating, Comment: d.Comment, Tags: tags, Replies: replies,
		HelpfulCount: d.HelpfulCount, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

type reviewMongoRepository struct{ coll *mongo.Collection }

func NewReviewRepository(db *mongo.Database) ReviewRepository {
	return &reviewMongoRepository{coll: db.Collection("course_reviews")}
}

// now membulatkan waktu ke milidetik. Tipe date pada BSON hanya menyimpan
// sampai milidetik, sedangkan time.Time di Go sampai nanodetik. Tanpa
// pembulatan, waktu pada response POST berbeda dari waktu saat dibaca kembali.
func now() time.Time { return time.Now().UTC().Truncate(time.Millisecond) }

func parseID(id string) (bson.ObjectID, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return bson.ObjectID{}, ErrInvalidID
	}
	return oid, nil
}

// Stub sementara; isinya diganti di Step 7.
func (r *reviewMongoRepository) EnsureIndexes(ctx context.Context) error { return nil }
