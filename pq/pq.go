package pq

import (
	//"container/heap"
	v1alpha1 "monitoring/api/v1"
	"time"
)

type Item struct {
	value    []*v1alpha1.Metric
	priority time.Duration
	index    int
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Greater(i, j int) bool {
	return pq[i].priority <= pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	temp := pq[i]
	pq[i] = pq[j]
	pq[j] = temp
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x *Item) {
	n := len(*pq)
	x.index = n
	*pq = append(*pq, x)
}

func (pq *PriorityQueue) Pop() *Item {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}
