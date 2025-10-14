package repositories

import (
	"errors"
	"fmt"
	"strings"

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
	return r.DeleteStringTx(key, r.db)
}

func (r *KeyValueRepository) DeleteStringTx(key string, tx *gorm.DB) error {
	result := tx.Delete(&models.KeyStringValue{}, &models.KeyStringValue{Key: key})

	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected != 1 {
		return errors.New("nothing deleted")
	}

	return nil
}

func (r *KeyValueRepository) DeleteWildcard(pattern string) error {
	return r.DeleteWildcardTx(pattern, r.db)
}

func (r *KeyValueRepository) DeleteWildcardTx(pattern string, tx *gorm.DB) error {
	return tx.
		Where(utils.QuoteSql(r.db, "%s like ?", "key"), strings.ReplaceAll(pattern, "*", "%")).
		Delete(&models.KeyStringValue{}).Error
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
