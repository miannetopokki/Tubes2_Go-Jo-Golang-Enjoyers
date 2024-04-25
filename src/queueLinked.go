package main

// ListNode represents a node in the linked list
type ListNode[T any] struct {
    value T
    next  *ListNode[T]
}

// Queue represents a queue implemented using a linked list
type QueueLinked[T any] struct {
    head *ListNode[T]
    tail *ListNode[T]
}

// Enqueue adds an element to the end of the queue
func (q *QueueLinked[T]) Enqueue(item T) {
    newNode := &ListNode[T]{value: item, next: nil}

    if q.tail == nil {
        // If queue is empty, set both head and tail to the new node
        q.head = newNode
        q.tail = newNode
    } else {
        // Append the new node to the end of the queue
        q.tail.next = newNode
        q.tail = newNode
    }
}

// EnqueueHead adds an element to the front of the queue
func (q *QueueLinked[T]) EnqueueHead(item T) {
    newNode := &ListNode[T]{value: item, next: q.head}

    if q.head == nil {
        // If queue is empty, set both head and tail to the new node
        q.head = newNode
        q.tail = newNode
    } else {
        // Update the head to point to the new node
        q.head = newNode
    }
}

// Dequeue removes and returns the element at the front of the queue
func (q *QueueLinked[T]) Dequeue() T {
    if q.head == nil {
        panic("Queue is empty")
    }

    // Get the value of the node at the head of the queue
    item := q.head.value

    // Move the head pointer to the next node
    q.head = q.head.next

    // If head becomes nil (queue becomes empty), also set tail to nil
    if q.head == nil {
        q.tail = nil
    }

    return item
}

// IsEmpty returns true if the queue is empty
func (q *QueueLinked[T]) IsEmpty() bool {
    return q.head == nil
}