package commonutils

import "fmt"

// Pair хранит упорядоченную пару значений обобщённых типов.
// T1 и T2 могут быть разными типами; ограничение comparable нужно для Equals и использования в map/set.
type Pair[T1 comparable, T2 comparable] struct {
	First  T1
	Second T2
}

// NewPair создаёт пару.
func NewPair[T1 comparable, T2 comparable](first T1, second T2) Pair[T1, T2] {
	return Pair[T1, T2]{First: first, Second: second}
}

// String форматирует пару для логов и печати.
func (p Pair[T1, T2]) String() string {
	return fmt.Sprintf("(%v, %v)", p.First, p.Second)
}

// Equals сравнивает пары по значению.
func (p Pair[T1, T2]) Equals(other Pair[T1, T2]) bool {
	return p.First == other.First && p.Second == other.Second
}

// IsZero проверяет, что оба значения равны нулевому значению своих типов.
func (p Pair[T1, T2]) IsZero() bool {
	var z1 T1
	var z2 T2
	return p.First == z1 && p.Second == z2
}

// Unpack удобная распаковка в два значения.
func (p Pair[T1, T2]) Unpack() (T1, T2) {
	return p.First, p.Second
}

// PairString — удобный алиас для пары строк.
type PairString = Pair[string, string]

// NewStringPair создаёт пару строк.
func NewStringPair(first, second string) PairString {
	return NewPair[string, string](first, second)
}
