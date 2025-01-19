package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.63

import (
	"context"

	"github.com/photoview/photoview/api/graphql/models"
	"gorm.io/gorm"
)

// SetThumbnailDownsampleMethod is the resolver for the setThumbnailDownsampleMethod field.
func (r *mutationResolver) SetThumbnailDownsampleMethod(ctx context.Context, method models.ThumbnailFilter) (models.ThumbnailFilter, error) {
	db := r.DB(ctx)

	// if method > 5 {
	// 	return 0, errors.New("The requested filter is unsupported, defaulting to nearest neighbor")
	// }

	if err := db.
		Session(&gorm.Session{AllowGlobalUpdate: true}).
		Model(&models.SiteInfo{}).
		Update("thumbnail_method", method).
		Error; err != nil {

		return models.ThumbnailFilterNearestNeighbor, err
	}

	var siteInfo models.SiteInfo
	if err := db.First(&siteInfo).Error; err != nil {
		return models.ThumbnailFilterNearestNeighbor, err
	}

	return siteInfo.ThumbnailMethod, nil

	// var langTrans *models.LanguageTranslation = nil
	// if language != nil {
	// 	lng := models.LanguageTranslation(*language)
	// 	langTrans = &lng
	// }
}
