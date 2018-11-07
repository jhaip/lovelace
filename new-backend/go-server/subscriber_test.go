package main

import (
	"reflect"
	"testing"

	"github.com/alecthomas/repr"
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
	expected_nodes["*query0 A B"] = makeNodeFromVariableNames([]string{"*query0", "A", "B"})
	expected_nodes["*query1 B C"] = makeNodeFromVariableNames([]string{"*query1", "B", "C"})
	expected_nodes["*query0 *query1 A B C"] = makeNodeFromVariableNames([]string{"*query0", "*query1", "A", "B", "C"})
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
	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	claim := []Term{Term{"text", "Sun"}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription = subscriberClaimUpdate(subscription, claim)

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	claim2 := []Term{Term{"text", "yellow"}, Term{"text", "has"}, Term{"integer", "3"}}
	subscription = subscriberClaimUpdate(subscription, claim2)

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	claim3 := []Term{Term{"text", "yellow"}, Term{"text", "has"}, Term{"text", "feelings"}}
	subscription = subscriberClaimUpdate(subscription, claim3)

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	claim4 := []Term{Term{"text", "Sky"}, Term{"text", "is"}, Term{"text", "blue"}}
	subscription = subscriberClaimUpdate(subscription, claim4)

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	retract1 := []Term{Term{"variable", ""}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription = subscriberRetractUpdate(subscription, retract1)

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))
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
	expected_nodes["*query0 A B"] = makeNodeFromVariableNames([]string{"*query0", "A", "B"})
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
	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	claim := []Term{Term{"text", "Sun"}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription = subscriberClaimUpdate(subscription, claim)

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	// claim does not match
	claim2 := []Term{Term{"text", "Sun"}, Term{"text", "BAD"}, Term{"text", "yellow"}}
	subscription = subscriberClaimUpdate(subscription, claim2)

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	// $ is BAD does not match any previous claim, so this should do nothing
	retract1 := []Term{Term{"variable", ""}, Term{"text", "is"}, Term{"text", "BAD"}}
	subscription = subscriberRetractUpdate(subscription, retract1)

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	// $ is yellow matches a previous claim "Sun is yellow" so that claim should be removed
	retract2 := []Term{Term{"variable", ""}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription = subscriberRetractUpdate(subscription, retract2)

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))
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
	expected_nodes["*query0"] = makeNodeFromVariableNames([]string{"*query0"})
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
	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	claim := []Term{Term{"text", "Sun"}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription = subscriberClaimUpdate(subscription, claim)

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))
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
	expected_nodes["*query0"] = makeNodeFromVariableNames([]string{"*query0"})
	expected_nodes["*query1"] = makeNodeFromVariableNames([]string{"*query1"})
	expected_nodes["*query0 *query1"] = makeNodeFromVariableNames([]string{"*query0", "*query1"})
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
	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	claim := []Term{Term{"text", "Sun"}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription = subscriberClaimUpdate(subscription, claim)

	claim2 := []Term{Term{"text", "Earth"}, Term{"text", "is"}, Term{"text", "green"}}
	subscription = subscriberClaimUpdate(subscription, claim2)

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))
}
