package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMakeSubscriber1(t *testing.T) {
	query_part1 := []Term{
		Term{"variable", "A"},
		Term{"text", "is"},
		Term{"variable", "B"},
	}
	query_part2 := []Term{
		Term{"variable", "B"},
		Term{"text", "has"},
		Term{"variable", "C"},
	}
	query := [][]Term{query_part1, query_part2}
	subscription := makeSubscriber("mySource", "0001", query)
	if len(subscription.nodes) != 3 {
		t.Error("Wrong number of nodes ", len(subscription.nodes))
	}
	expected_nodes := make(map[string]Node)
	expected_nodes["*query0 A B"] = Node{[]string{"*query0", "A", "B"}, make([]NodeValue, 0)}
	expected_nodes["*query1 B C"] = Node{[]string{"*query1", "B", "C"}, make([]NodeValue, 0)}
	expected_nodes["*query0 *query1 A B C"] = Node{[]string{"*query0", "*query1", "A", "B", "C"}, make([]NodeValue, 0)}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	if len(subscription.queryPartToUpdate) != len(query) {
		t.Error("length of queryPartToUpdate should match query len", len(subscription.queryPartToUpdate), len(query))
	}
	expected_queryPartToUpdate0 := []SubscriptionUpdateOptions{
		SubscriptionUpdateOptions{"*query1 B C", "*query0 *query1 A B C", []string{"*query0", "A"}}}
	expected_queryPartToUpdate1 := []SubscriptionUpdateOptions{
		SubscriptionUpdateOptions{"*query0 A B", "*query0 *query1 A B C", []string{"*query1", "C"}}}
	if !reflect.DeepEqual(subscription.queryPartToUpdate[0], expected_queryPartToUpdate0) {
		t.Error("subscription.queryPartToUpdate[0] does not match")
	}
	if !reflect.DeepEqual(subscription.queryPartToUpdate[1], expected_queryPartToUpdate1) {
		t.Error("subscription.queryPartToUpdate[1] does not match")
	}
	fmt.Println(subscription)
}

func TestMakeSubscriberOnePart(t *testing.T) {
	query_part1 := []Term{
		Term{"variable", "A"},
		Term{"text", "is"},
		Term{"variable", "B"},
	}
	query := [][]Term{query_part1}
	subscription := makeSubscriber("mySource", "0001", query)
	if len(subscription.nodes) != 1 {
		t.Error("Wrong number of nodes ", len(subscription.nodes))
	}
	expected_nodes := make(map[string]Node)
	expected_nodes["*query0 A B"] = Node{[]string{"*query0", "A", "B"}, make([]NodeValue, 0)}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}
	if len(subscription.queryPartToUpdate) != len(query) {
		t.Error("length of queryPartToUpdate should match query len", len(subscription.queryPartToUpdate), len(query))
	}
	expected_queryPartToUpdate0 := make([]SubscriptionUpdateOptions, 0)
	if !reflect.DeepEqual(subscription.queryPartToUpdate[0], expected_queryPartToUpdate0) {
		t.Error("subscription.queryPartToUpdate[0] does not match")
	}
	fmt.Println(subscription)
}

func TestMakeSubscriberOnePartNoVariables(t *testing.T) {
	query_part1 := []Term{
		Term{"text", "Sun"},
		Term{"text", "is"},
		Term{"text", "yellow"},
	}
	query := [][]Term{query_part1}
	subscription := makeSubscriber("mySource", "0001", query)
	if len(subscription.nodes) != 1 {
		t.Error("Wrong number of nodes ", len(subscription.nodes))
	}
	expected_nodes := make(map[string]Node)
	expected_nodes["*query0"] = Node{[]string{"*query0"}, make([]NodeValue, 0)}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}
	if len(subscription.queryPartToUpdate) != len(query) {
		t.Error("length of queryPartToUpdate should match query len", len(subscription.queryPartToUpdate), len(query))
	}
	expected_queryPartToUpdate0 := make([]SubscriptionUpdateOptions, 0)
	if !reflect.DeepEqual(subscription.queryPartToUpdate[0], expected_queryPartToUpdate0) {
		t.Error("subscription.queryPartToUpdate[0] does not match")
	}
	fmt.Println(subscription)
}

func TestMakeSubscriberTwoPartsNoVariables(t *testing.T) {
	query_part1 := []Term{
		Term{"text", "Sun"},
		Term{"text", "is"},
		Term{"text", "yellow"},
	}
	query_part2 := []Term{
		Term{"text", "Earth"},
		Term{"text", "is"},
		Term{"text", "green"},
	}
	query := [][]Term{query_part1, query_part2}
	subscription := makeSubscriber("mySource", "0001", query)
	if len(subscription.nodes) != 3 {
		t.Error("Wrong number of nodes ", len(subscription.nodes))
	}
	expected_nodes := make(map[string]Node)
	expected_nodes["*query0"] = Node{[]string{"*query0"}, make([]NodeValue, 0)}
	expected_nodes["*query1"] = Node{[]string{"*query1"}, make([]NodeValue, 0)}
	expected_nodes["*query0 *query1"] = Node{[]string{"*query0", "*query1"}, make([]NodeValue, 0)}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	if len(subscription.queryPartToUpdate) != len(query) {
		t.Error("length of queryPartToUpdate should match query len", len(subscription.queryPartToUpdate), len(query))
	}
	expected_queryPartToUpdate0 := []SubscriptionUpdateOptions{
		SubscriptionUpdateOptions{"*query1", "*query0 *query1", []string{"*query0"}}}
	expected_queryPartToUpdate1 := []SubscriptionUpdateOptions{
		SubscriptionUpdateOptions{"*query0", "*query0 *query1", []string{"*query1"}}}
	if !reflect.DeepEqual(subscription.queryPartToUpdate[0], expected_queryPartToUpdate0) {
		t.Error("subscription.queryPartToUpdate[0] does not match")
	}
	if !reflect.DeepEqual(subscription.queryPartToUpdate[1], expected_queryPartToUpdate1) {
		t.Error("subscription.queryPartToUpdate[1] does not match")
	}
	fmt.Println(subscription)
}
