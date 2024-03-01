package deep_copy

import (
	"testing"
)

type TestStruct struct {
	a             int
	MapInt64Int64 map[int64]int64
	Int64         int64
	ArrayInt64    [4]int64
	SliceInt64    []int64
}

func NewTestStruct() TestStruct {
	v := TestStruct{}
	v.a = 121
	v.MapInt64Int64 = make(map[int64]int64)
	v.SliceInt64 = make([]int64, 10)

	v.MapInt64Int64[2] = 2
	v.ArrayInt64[3] = 3
	v.SliceInt64[4] = 4

	return v
}

func TestDeepCopy(t *testing.T) {
	original := NewTestStruct()
	copied := MustCopy(original).(TestStruct)

	copied.MapInt64Int64[2] = 3
	copied.ArrayInt64[3] = 4
	original.SliceInt64[4] = 5

	failed := false

	failed = failed || (original.MapInt64Int64[2] == copied.MapInt64Int64[2])
	failed = failed || (original.ArrayInt64[3] == copied.ArrayInt64[3])
	failed = failed || (original.SliceInt64[4] == copied.SliceInt64[4])

	if failed {
		t.Fatalf(`copy failed`)
	}
}

func BenchmarkDeepCopy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v1 := NewTestStruct()
		v2 := MustCopy(&v1)
		_ = v2
	}
}
