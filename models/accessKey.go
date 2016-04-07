package models

import "github.com/ansible-semaphore/semaphore/database"

type AccessKey struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name" binding:"required"`
	// 'aws/do/gcloud/ssh',
	Type string `db:"type" json:"type" binding:"required"`

	ProjectID *int    `db:"project_id" json:"project_id"`
	Key       *string `db:"key" json:"key"`
	Secret    *string `db:"secret" json:"secret"`
}

func init() {
	database.Mysql.AddTableWithName(AccessKey{}, "access_key").SetKeys(true, "id")
}
