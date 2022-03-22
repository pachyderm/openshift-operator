package api

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Metadata of the records
type Metadata struct {
	// ID of the record
	ID uuid.UUID `json:"id,omitempty"`
	// Name of object
	Name string `json:"name,omitempty"`
	// Namespace of the object
	Namespace string `json:"namespace,omitempty"`
	// CreatedAt defines time record was created
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// UpdatedAt captures time record is updated
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	// DeletedAt captures time record is updated
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// Backup resource
type Backup struct {
	Metadata      `json:"metadata,omitempty"`
	IsRunning     bool     `json:"is_running,omitempty"`
	PodName       string   `json:"pod_name,omitempty"`
	ContainerName string   `json:"container_name,omitempty"`
	Command       []string `json:"command,omitempty"`
	// cmd is used internally to store command as string
	Cmd string `json:"-"`
}

// BackupList response object
type BackupList struct {
	Items []*Backup `json:"items,omitempty"`
}

func (b *Backup) New() {
	b.Metadata.CreatedAt = utcTime()
	b.Metadata.ID = uuid.New()
}

func (b *Backup) SetUpdatedTime() {
	b.Metadata.UpdatedAt = utcTime()
}

func (b *Backup) SetDeletedTime() {
	b.Metadata.DeletedAt = utcTime()
}

func utcTime() *time.Time {
	utc := time.Now().UTC()
	return &utc
}

// Takes the Backup.Command value of time []string
// and stores it to Backup.cmd as a base64 encoded string.
func (b *Backup) EncodeCmd() {
	var tmp []string
	for _, s := range b.Command {
		tmp = append(tmp, base64.StdEncoding.EncodeToString([]byte(s)))
	}
	b.Cmd = strings.Join(tmp, ".")
}

// Decodes the contents of Backup.cmd and stores it to Backup.Command of type
// []string
func (b *Backup) DecodeCmd() {
	b.Command = []string{}
	for _, s := range strings.Split(b.Cmd, ".") {
		payload, _ := base64.StdEncoding.DecodeString(s)
		b.Command = append(b.Command, string(payload))
	}
}
