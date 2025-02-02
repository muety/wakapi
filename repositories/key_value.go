package repositories

import (
	"errors"
	"fmt"

	"github.com/muety/wakapi/models"
	"github.com/muety/wakapi/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type KeyValueRepository struct {
	BaseRepository
}

func NewKeyValueRepository(db *gorm.DB) *KeyValueRepository {
	return &KeyValueRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *KeyValueRepository) GetAll() ([]*models.KeyStringValue, error) {
	var keyValues []*models.KeyStringValue
	if err := r.db.Find(&keyValues).Error; err != nil {
		return nil, err
	}
	return keyValues, nil
}

func (r *KeyValueRepository) GetString(key string) (*models.KeyStringValue, error) {
	kv := &models.KeyStringValue{}
	if err := r.db.
		Where(&models.KeyStringValue{Key: key}).
		First(&kv).Error; err != nil {
		return nil, err
	}

	return kv, nil
}

func (r *KeyValueRepository) Search(like string) ([]*models.KeyStringValue, error) {
	var keyValues []*models.KeyStringValue
	if err := r.db.Table("key_string_values").
		Where(utils.QuoteSql(r.db, "%s like ?", "key"), like).
		Find(&keyValues).
		Error; err != nil {
		return nil, err
	}
	return keyValues, nil
}

func (r *KeyValueRepository) PutString(kv *models.KeyStringValue) error {
	result := r.db.
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		Where(&models.KeyStringValue{Key: kv.Key}).
		Assign(kv).
		Create(kv)

	if err := result.Error; err != nil {
		return err
	}

	return nil
}

func (r *KeyValueRepository) DeleteString(key string) error {
	result := r.db.
		Delete(&models.KeyStringValue{}, &models.KeyStringValue{Key: key})

	if err := result.Error; err != nil {
		return err
	}

	if result.RowsAffected != 1 {
		return errors.New("nothing deleted")
	}

	return nil
}

// ReplaceKeySuffix will search for key-value pairs whose key ends with suffixOld and replace it with suffixNew instead.
func (r *KeyValueRepository) ReplaceKeySuffix(suffixOld, suffixNew string) error {
	if dialector := r.db.Dialector.Name(); dialector == "mysql" || dialector == "postgres" {
		patternOld := fmt.Sprintf("(.+)%s$", suffixOld)
		patternNew := fmt.Sprintf("$1%s", suffixNew) // mysql group replace style
		if dialector == "postgres" {
			patternNew = fmt.Sprintf("\\1%s", suffixNew) // postgres group replace style
		}

		return r.db.Model(&models.KeyStringValue{}).
			Where(utils.QuoteSql(r.db, "%s like ?", "key"), "%"+suffixOld).
			Update("key", gorm.Expr(
				utils.QuoteSql(r.db, "regexp_replace(%s, ?, ?)", "key"),
				patternOld,
				patternNew,
			)).Error
	} else {
		// a bit less safe, because not only replacing suffixes
		return r.db.Model(&models.KeyStringValue{}).
			Where("key like ?", "%"+suffixOld).
			Update("key", gorm.Expr("replace(key, ?, ?)", suffixOld, suffixNew)).Error
	}
}
