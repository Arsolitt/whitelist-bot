package utils

import "github.com/google/uuid"

// UUIDBasedID - constraint для всех типов, базирующихся на uuid.UUID
type UUIDBasedID interface {
	~[16]byte // uuid.UUID это [16]byte под капотом
}

// UUIDString - generic функция для преобразования UUID-based типов в строку
func UUIDString[T UUIDBasedID](id T) string {
	return uuid.UUID(id).String()
}

// UUIDIsZero - generic функция для проверки, является ли UUID нулевым
func UUIDIsZero[T UUIDBasedID](id T) bool {
	uuidVal := uuid.UUID(id)
	return uuidVal == uuid.Nil || uuidVal.String() == ""
}
