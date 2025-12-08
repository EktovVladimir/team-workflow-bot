package common

import "fmt"

// Triple хранит упорядоченную тройку значений обобщённых типов.
// Название Triple наиболее ожидаемое; альтернативы: Trio, Triplet, Ternary.
type Triple[T1 comparable, T2 comparable, T3 comparable] struct {
	First  T1
	Second T2
	Third  T3
}

// NewTriple создаёт тройку.
func NewTriple[T1 comparable, T2 comparable, T3 comparable](first T1, second T2, third T3) Triple[T1, T2, T3] {
	return Triple[T1, T2, T3]{First: first, Second: second, Third: third}
}

// String форматирует тройку для логов и печати.
func (t Triple[T1, T2, T3]) String() string {
	return fmt.Sprintf("(%v, %v, %v)", t.First, t.Second, t.Third)
}

// Equals сравнивает тройки по значению.
func (t Triple[T1, T2, T3]) Equals(other Triple[T1, T2, T3]) bool {
	return t.First == other.First && t.Second == other.Second && t.Third == other.Third
}

// IsZero проверяет, что все значения равны нулевому значению своих типов.
func (t Triple[T1, T2, T3]) IsZero() bool {
	var z1 T1
	var z2 T2
	var z3 T3
	return t.First == z1 && t.Second == z2 && t.Third == z3
}

// Unpack удобная распаковка в три значения.
func (t Triple[T1, T2, T3]) Unpack() (T1, T2, T3) {
	return t.First, t.Second, t.Third
}
