package memoize

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func AssertEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	t.Errorf("Received %v (type %v), expected %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

func TestPanic(t *testing.T) {
	count := 0
	f := func(i int) {
		count++
		if count%2 == 1 {
			panic(count)
		}
	}
	f = Memoize(f).(func(int))

	expect := func(p interface{}, i int) {
		defer func() {
			if r := recover(); p != r {
				t.Errorf("for input %d:\nexpected: %v\nactual: %v", i, p, r)
			}
		}()

		f(i)
	}

	expect(1, 1)
	expect(1, 1)
	expect(nil, 2)
	expect(nil, 2)
	expect(1, 1)
	expect(3, 100)
}

func TestVariadic(t *testing.T) {
	count := 0
	var concat func(string, ...string) string
	concat = func(s0 string, s1 ...string) string {
		count++

		if len(s1) == 0 {
			return s0
		}
		return concat(s0+s1[0], s1[1:]...)
	}
	concat = Memoize(concat).(func(string, ...string) string)

	expect := func(actual, expected string, n int) {
		if actual != expected || n != count {
			t.Errorf("expected: %q\nactual: %q\nexpected count: %d\nactual count: %d", expected, actual, n, count)
		}
	}

	expect("", "", 0)
	expect(concat("string"), "string", 1)
	expect(concat("string", "one"), "stringone", 3)
	expect(concat("string", "one"), "stringone", 3)
	expect(concat("string", "two"), "stringtwo", 5)
	expect(concat("string", "one"), "stringone", 5)
	expect(concat("stringone", "two"), "stringonetwo", 7)
	expect(concat("string", "one", "two"), "stringonetwo", 8)
}

func TestMixed(t *testing.T) {
	var foo func(int, string, float64) (int, string)
	foo = func(i int, s0 string, f0 float64) (int, string) {
		fmt.Println("running", s0)
		time.Sleep(time.Second * time.Duration(f0))
		out := fmt.Sprintf("%d %s %.1f", i, s0, f0)
		return i + 1, out
	}
	foo = Memoize(foo).(func(int, string, float64) (int, string))

	i, s := foo(0, "delay", 3.2)
	AssertEqual(t, i, 1)
	AssertEqual(t, s, "0 delay 3.2")

	i, s = foo(1, "nodelay", 0)
	AssertEqual(t, i, 2)
	AssertEqual(t, s, "1 nodelay 0.0")

	i, s = foo(0, "delay", 3.2)
	AssertEqual(t, i, 1)
	AssertEqual(t, s, "0 delay 3.2")

	i, s = foo(1, "nodelay", 0)
	AssertEqual(t, i, 2)
	AssertEqual(t, s, "1 nodelay 0.0")

}
