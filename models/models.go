package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type BaseModel struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type Queue struct {
	BaseModel
	TaskID     string `json:"task_id"`
	SentToPeon bool   `json:"queued"`
}

type Task struct {
	BaseModel
	TaskName       string      `json:"task_name"`
	Status         TaskStatus  `json:"status"`
	PeonID         *string     `json:"peon_id"`
	Queue          string      `json:"queue"`
	PayloadStr     string      `json:"-"`
	ResultStr      string      `json:"-"`
	RetryOnFailure bool        `json:"retry_on_failure"`
	RetryCount     int         `json:"retry_count"`
	RetryLimit     int         `json:"retry_limit"`
	Payload        TaskPayload `json:"payload"`
	Result         interface{} `json:"result"`
	Logs           *string     `json:"logs"`
}

type Stats struct {
	BaseModel
	Type     string      `json:"type"`
	ValueStr string      `json:"-"` // Hide from JSON
	PeonID   *string     `json:"peon_id"`
	TaskID   *string     `json:"task_id"`
	Value    interface{} `json:"value"` // Use this for JSON
}

type Peon struct {
	BaseModel
	Status        string  `json:"status"`
	LastHeartbeat string  `json:"last_heartbeat"`
	CurrentTask   *string `json:"current_task"`
	Queues        *string `json:"queues"`
}

type PeonUpdate struct {
	Status         *string `json:"status,omitempty"`
	StatusSet      bool    `json:"-"`
	LastHeartbeat  *string `json:"last_heartbeat,omitempty"`
	HeartbeatSet   bool    `json:"-"`
	CurrentTask    *string `json:"current_task,omitempty"`
	CurrentTaskSet bool    `json:"-"`
	Queues         *string `json:"queues,omitempty"`
	QueuesSet      bool    `json:"-"`
}

type TaskUpdate struct {
	Status            *string      `json:"status,omitempty"`
	StatusSet         bool         `json:"-"`
	TaskName          *string      `json:"task_name,omitempty"`
	TaskNameSet       bool         `json:"-"`
	PeonID            *string      `json:"peon_id,omitempty"`
	PeonIDSet         bool         `json:"-"`
	Queue             *string      `json:"queue,omitempty"`
	QueueSet          bool         `json:"-"`
	Payload           *interface{} `json:"payload,omitempty"`
	PayloadSet        bool         `json:"-"`
	Result            *string      `json:"result,omitempty"`
	ResultSet         bool         `json:"-"`
	RetryOnFailure    *bool        `json:"retry_on_failure,omitempty"`
	RetryOnFailureSet bool         `json:"-"`
	RetryCount        *int         `json:"retry_count,omitempty"`
	RetryCountSet     bool         `json:"-"`
	RetryLimit        *int         `json:"retry_limit,omitempty"`
	RetryLimitSet     bool         `json:"-"`
	Logs              *string      `json:"logs,omitempty"`
	LogsSet           bool         `json:"-"`
}

type SSEMessage struct {
	Type    string                  `json:"type"`
	Message *map[string]interface{} `json:"message,omitempty"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	jwt.RegisteredClaims
	APIKey string `json:"api_key"`
}

type TaskAcknowledgement struct {
	PeonID string `json:"peon_id"`
}

type TaskStatus string

const (
	TaskStatusPending      TaskStatus = "PENDING"
	TaskStatusRunning      TaskStatus = "RUNNING"
	TaskStatusSuccess      TaskStatus = "SUCCESS"
	TaskStatusFailure      TaskStatus = "FAILURE"
	TaskStatusInvalid      TaskStatus = "INVALID"
	TaskStatusCancelled    TaskStatus = "CANCELLED"
	TaskStatusAcknowledged TaskStatus = "ACKNOWLEDGED"
)

type TaskPayload struct {
	TaskArgs             []interface{}          `json:"task_args"`
	TaskKwargs           map[string]interface{} `json:"task_kwargs"`
	PrerunHandlerArgs    []interface{}          `json:"prerun_handler_args"`
	PrerunHandlerKwargs  map[string]interface{} `json:"prerun_handler_kwargs"`
	PostrunHandlerArgs   []interface{}          `json:"postrun_handler_args"`
	PostrunHandlerKwargs map[string]interface{} `json:"postrun_handler_kwargs"`
}

type FilterOperator string

const (
	FilterOpEquals    FilterOperator = "eq"
	FilterOpGreater   FilterOperator = "gt"
	FilterOpLess      FilterOperator = "lt"
	FilterOpGreaterEq FilterOperator = "gte"
	FilterOpLessEq    FilterOperator = "lte"
	FilterOpIn        FilterOperator = "in"
	FilterOpNotIn     FilterOperator = "not_in"
)

type FilterCondition struct {
	Op    FilterOperator `json:"op"`
	Value interface{}    `json:"value"`
}

type QueryParams struct {
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
	Order   *OrderParam `json:"order,omitempty"`
	Filter  any         `json:"filter,omitempty"` // will be TaskFilter or PeonFilter
}

type OrderParam struct {
	Field string `json:"field"` // e.g., "created_at", "last_heartbeat"
	Dir   string `json:"dir"`   // "ASC" or "DESC"
}

// For Tasks
type TaskQuery struct {
	QueryParams
	Filter *TaskFilter `json:"filter,omitempty"`
}

// For Peons
type PeonQuery struct {
	QueryParams
	Filter *PeonFilter `json:"filter,omitempty"`
}

type TaskFilter struct {
	Status    *FilterCondition `json:"status,omitempty"`
	CreatedAt *FilterCondition `json:"created_at,omitempty"`
	TaskName  *FilterCondition `json:"task_name,omitempty"`
	Queue     *FilterCondition `json:"queue,omitempty"`
	PeonID    *FilterCondition `json:"peon_id,omitempty"`
}

type PeonFilter struct {
	Status        *FilterCondition `json:"status,omitempty"`
	LastHeartbeat *FilterCondition `json:"last_heartbeat,omitempty"`
	CurrentTask   *FilterCondition `json:"current_task,omitempty"`
	Queues        *FilterCondition `json:"queues,omitempty"`
}

type PaginatedResponse struct {
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalItems int         `json:"total_items"`
	TotalPages int         `json:"total_pages"`
	Items      interface{} `json:"items"`
}
