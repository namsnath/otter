package db

type Creator[T any] interface {
	Create(T) T
}

type Updater[T any] interface {
	Update(T) T
}

type Deleter[T any] interface {
	Delete(T) T
}

type ChildCreator[T any] interface {
	CreateAsChildOf(T) T
}
