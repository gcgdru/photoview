package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.63

import (
	"context"
	"errors"

	api "github.com/photoview/photoview/api/graphql"
	"github.com/photoview/photoview/api/graphql/auth"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/scanner/face_detection"
	"gorm.io/gorm"
)

// ImageFaces is the resolver for the imageFaces field.
func (r *faceGroupResolver) ImageFaces(ctx context.Context, obj *models.FaceGroup, paginate *models.Pagination) ([]*models.ImageFace, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	if err := user.FillAlbums(db); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	query := db.
		Joins("Media").
		Where(faceGroupIDIsQuestion, obj.ID).
		Where("album_id IN (?)", userAlbumIDs)

	query = models.FormatSQL(query, nil, paginate)

	var imageFaces []*models.ImageFace
	if err := query.Find(&imageFaces).Error; err != nil {
		return nil, err
	}

	return imageFaces, nil
}

// ImageFaceCount is the resolver for the imageFaceCount field.
func (r *faceGroupResolver) ImageFaceCount(ctx context.Context, obj *models.FaceGroup) (int, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return -1, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return -1, ErrFaceDetectorNotInitialized
	}

	if err := user.FillAlbums(db); err != nil {
		return -1, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	query := db.
		Model(&models.ImageFace{}).
		Joins("Media").
		Where(faceGroupIDIsQuestion, obj.ID).
		Where("album_id IN (?)", userAlbumIDs)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return -1, err
	}

	return int(count), nil
}

// Media is the resolver for the media field.
func (r *imageFaceResolver) Media(ctx context.Context, obj *models.ImageFace) (*models.Media, error) {
	if err := obj.FillMedia(r.DB(ctx)); err != nil {
		return nil, err
	}

	return &obj.Media, nil
}

// FaceGroup is the resolver for the faceGroup field.
func (r *imageFaceResolver) FaceGroup(ctx context.Context, obj *models.ImageFace) (*models.FaceGroup, error) {
	if obj.FaceGroup != nil {
		return obj.FaceGroup, nil
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	var faceGroup models.FaceGroup
	if err := r.DB(ctx).Model(&obj).Association("FaceGroup").Find(&faceGroup); err != nil {
		return nil, err
	}

	obj.FaceGroup = &faceGroup

	return &faceGroup, nil
}

// SetFaceGroupLabel is the resolver for the setFaceGroupLabel field.
func (r *mutationResolver) SetFaceGroupLabel(ctx context.Context, faceGroupID int, label *string) (*models.FaceGroup, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	faceGroup, err := userOwnedFaceGroup(db, user, faceGroupID)
	if err != nil {
		return nil, err
	}

	if err := db.Model(faceGroup).Update("label", label).Error; err != nil {
		return nil, err
	}

	return faceGroup, nil
}

// CombineFaceGroups is the resolver for the combineFaceGroups field.
func (r *mutationResolver) CombineFaceGroups(ctx context.Context, destinationFaceGroupID int, sourceFaceGroupID int) (*models.FaceGroup, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	destinationFaceGroup, err := userOwnedFaceGroup(db, user, destinationFaceGroupID)
	if err != nil {
		return nil, err
	}

	sourceFaceGroup, err := userOwnedFaceGroup(db, user, sourceFaceGroupID)
	if err != nil {
		return nil, err
	}

	updateError := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ImageFace{}).
			Where(faceGroupIDIsQuestion, sourceFaceGroup.ID).
			Update("face_group_id", destinationFaceGroup.ID).Error; err != nil {
			return err
		}

		if err := tx.Delete(&sourceFaceGroup).Error; err != nil {
			return err
		}

		return nil
	})

	if updateError != nil {
		return nil, updateError
	}

	face_detection.GlobalFaceDetector.MergeCategories(int32(sourceFaceGroupID), int32(destinationFaceGroupID))

	return destinationFaceGroup, nil
}

// MoveImageFaces is the resolver for the moveImageFaces field.
func (r *mutationResolver) MoveImageFaces(ctx context.Context, imageFaceIDs []int, destinationFaceGroupID int) (*models.FaceGroup, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	userOwnedImageFaceIDs := make([]int, 0)
	var destFaceGroup *models.FaceGroup

	transErr := db.Transaction(func(tx *gorm.DB) error {

		var err error
		destFaceGroup, err = userOwnedFaceGroup(tx, user, destinationFaceGroupID)
		if err != nil {
			return err
		}

		userOwnedImageFaces, err := getUserOwnedImageFaces(tx, user, imageFaceIDs)
		if err != nil {
			return err
		}

		for _, imageFace := range userOwnedImageFaces {
			userOwnedImageFaceIDs = append(userOwnedImageFaceIDs, imageFace.ID)
		}

		var sourceFaceGroups []*models.FaceGroup
		if err := tx.
			Joins("LEFT JOIN image_faces ON image_faces.face_group_id = face_groups.id").
			Where(imageFacesIDInQuestion, userOwnedImageFaceIDs).
			Find(&sourceFaceGroups).Error; err != nil {
			return err
		}

		if err := tx.
			Model(&models.ImageFace{}).
			Where("id IN (?)", userOwnedImageFaceIDs).
			Update("face_group_id", destFaceGroup.ID).Error; err != nil {
			return err
		}

		// delete face groups if they have become empty
		if err := deleteEmptyFaceGroups(sourceFaceGroups, tx); err != nil {
			return err
		}

		return nil
	})

	if transErr != nil {
		return nil, transErr
	}

	face_detection.GlobalFaceDetector.MergeImageFaces(userOwnedImageFaceIDs, int32(destFaceGroup.ID))

	return destFaceGroup, nil
}

// RecognizeUnlabeledFaces is the resolver for the recognizeUnlabeledFaces field.
func (r *mutationResolver) RecognizeUnlabeledFaces(ctx context.Context) ([]*models.ImageFace, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	var updatedImageFaces []*models.ImageFace

	transactionError := db.Transaction(func(tx *gorm.DB) error {
		var err error
		updatedImageFaces, err = face_detection.GlobalFaceDetector.RecognizeUnlabeledFaces(tx, user)

		return err
	})

	if transactionError != nil {
		return nil, transactionError
	}

	return updatedImageFaces, nil
}

// DetachImageFaces is the resolver for the detachImageFaces field.
func (r *mutationResolver) DetachImageFaces(ctx context.Context, imageFaceIDs []int) (*models.FaceGroup, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	userOwnedImageFaceIDs := make([]int, 0)
	newFaceGroup := models.FaceGroup{}

	transactionError := db.Transaction(func(tx *gorm.DB) error {

		userOwnedImageFaces, err := getUserOwnedImageFaces(tx, user, imageFaceIDs)
		if err != nil {
			return err
		}

		for _, imageFace := range userOwnedImageFaces {
			userOwnedImageFaceIDs = append(userOwnedImageFaceIDs, imageFace.ID)
		}

		if err := tx.Save(&newFaceGroup).Error; err != nil {
			return err
		}

		if err := tx.
			Model(&models.ImageFace{}).
			Where("id IN (?)", userOwnedImageFaceIDs).
			Update("face_group_id", newFaceGroup.ID).Error; err != nil {
			return err
		}

		return nil
	})

	if transactionError != nil {
		return nil, transactionError
	}

	face_detection.GlobalFaceDetector.MergeImageFaces(userOwnedImageFaceIDs, int32(newFaceGroup.ID))

	return &newFaceGroup, nil
}

// MyFaceGroups is the resolver for the myFaceGroups field.
func (r *queryResolver) MyFaceGroups(ctx context.Context, paginate *models.Pagination) ([]*models.FaceGroup, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	if err := user.FillAlbums(db); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	faceGroupQuery := db.
		Joins("JOIN image_faces ON image_faces.face_group_id = face_groups.id").
		Where("image_faces.media_id IN (?)",
			db.Select("media.id").Table("media").Where(mediaAlbumIDInQuestion, userAlbumIDs)).
		Group("image_faces.face_group_id").
		Group("face_groups.id").
		Order("CASE WHEN label IS NULL THEN 1 ELSE 0 END").
		Order("COUNT(image_faces.id) DESC")

	faceGroupQuery = models.FormatSQL(faceGroupQuery, nil, paginate)

	var faceGroups []*models.FaceGroup
	if err := faceGroupQuery.Find(&faceGroups).Error; err != nil {
		return nil, err
	}

	return faceGroups, nil
}

// FaceGroup is the resolver for the faceGroup field.
func (r *queryResolver) FaceGroup(ctx context.Context, id int) (*models.FaceGroup, error) {
	db := r.DB(ctx)
	user := auth.UserFromContext(ctx)
	if user == nil {
		return nil, errors.New("unauthorized")
	}

	if face_detection.GlobalFaceDetector == nil {
		return nil, ErrFaceDetectorNotInitialized
	}

	if err := user.FillAlbums(db); err != nil {
		return nil, err
	}

	userAlbumIDs := make([]int, len(user.Albums))
	for i, album := range user.Albums {
		userAlbumIDs[i] = album.ID
	}

	faceGroupQuery := db.
		Joins("LEFT JOIN image_faces ON image_faces.face_group_id = face_groups.id").
		Joins("LEFT JOIN media ON image_faces.media_id = media.id").
		Where("face_groups.id = ?", id).
		Where(mediaAlbumIDInQuestion, userAlbumIDs)

	var faceGroup models.FaceGroup
	if err := faceGroupQuery.Find(&faceGroup).Error; err != nil {
		return nil, err
	}

	return &faceGroup, nil
}

// FaceGroup returns api.FaceGroupResolver implementation.
func (r *Resolver) FaceGroup() api.FaceGroupResolver { return &faceGroupResolver{r} }

// ImageFace returns api.ImageFaceResolver implementation.
func (r *Resolver) ImageFace() api.ImageFaceResolver { return &imageFaceResolver{r} }

type faceGroupResolver struct{ *Resolver }
type imageFaceResolver struct{ *Resolver }
