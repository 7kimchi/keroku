package gateway

import (
	"reflect"
	"testing"
)

func TestPlanUsesRecommendation(t *testing.T) {
	p, err := PlanShards(3, 0, nil, 1, 1000)
	if err != nil || p.Count != 3 || !reflect.DeepEqual(p.Batches, [][]int{{0}, {1}, {2}}) {
		t.Fatalf("got %+v %v", p, err)
	}
}

func TestPlanBatchesByConcurrency(t *testing.T) {
	p, err := PlanShards(1, 10, []int{9, 0, 4, 5, 1}, 2, 1000)
	if err != nil || p.Count != 10 || !reflect.DeepEqual(p.Batches, [][]int{{0, 1}, {4, 5}, {9}}) {
		t.Fatalf("got %+v %v", p, err)
	}
	p, err = PlanShards(16, 0, nil, 16, 1000)
	if err != nil || len(p.Batches) != 1 || len(p.Batches[0]) != 16 {
		t.Fatalf("got %+v %v", p, err)
	}
}

func TestPlanDoesNotMutateInput(t *testing.T) {
	ids := []int{3, 1}
	if _, err := PlanShards(1, 4, ids, 1, 10); err != nil || ids[0] != 3 {
		t.Fatal("input reordered")
	}
}

func TestPlanRejects(t *testing.T) {
	cases := []struct {
		rec, conf int
		ids       []int
		remaining int
	}{
		{0, 0, nil, 10},
		{1, 4, []int{4}, 10},
		{1, 4, []int{-1}, 10},
		{1, 4, []int{1, 1}, 10},
		{1, 4, []int{}, 10},
		{8, 0, nil, 7},
		{-5, 0, nil, 10},
	}
	for _, c := range cases {
		if _, err := PlanShards(c.rec, c.conf, c.ids, 1, c.remaining); err == nil {
			t.Errorf("%+v accepted", c)
		}
	}
}

func TestPlanZeroConcurrencyMeansOne(t *testing.T) {
	p, err := PlanShards(2, 0, nil, 0, 10)
	if err != nil || len(p.Batches) != 2 {
		t.Fatalf("got %+v %v", p, err)
	}
}
