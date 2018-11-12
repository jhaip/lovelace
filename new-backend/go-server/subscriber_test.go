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
	subscription := makeSubscriber(query)

	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))
	// return

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

	updatedSubscriberOutput := false

	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))
	claim := []Term{Term{"text", "Sun"}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim)
	if updatedSubscriberOutput != false {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0 A B"].variableCache["*query0"] = []NodeValue{NodeValue{claim}}
	expected_nodes["*query0 A B"].variableCache["A"] = []NodeValue{NodeValue{[]Term{Term{"text", "Sun"}}}}
	expected_nodes["*query0 A B"].variableCache["B"] = []NodeValue{NodeValue{[]Term{Term{"text", "yellow"}}}}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))
	claim2 := []Term{Term{"text", "yellow"}, Term{"text", "has"}, Term{"integer", "3"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim2)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query1 B C"].variableCache["*query1"] = []NodeValue{NodeValue{claim2}}
	expected_nodes["*query1 B C"].variableCache["B"] = []NodeValue{NodeValue{[]Term{Term{"text", "yellow"}}}}
	expected_nodes["*query1 B C"].variableCache["C"] = []NodeValue{NodeValue{[]Term{Term{"integer", "3"}}}}
	expected_nodes["*query0 *query1 A B C"].variableCache["*query0"] = []NodeValue{NodeValue{claim}}
	expected_nodes["*query0 *query1 A B C"].variableCache["*query1"] = []NodeValue{NodeValue{claim2}}
	expected_nodes["*query0 *query1 A B C"].variableCache["A"] = []NodeValue{NodeValue{[]Term{Term{"text", "Sun"}}}}
	expected_nodes["*query0 *query1 A B C"].variableCache["B"] = []NodeValue{NodeValue{[]Term{Term{"text", "yellow"}}}}
	expected_nodes["*query0 *query1 A B C"].variableCache["C"] = []NodeValue{NodeValue{[]Term{Term{"integer", "3"}}}}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}
	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	claim3 := []Term{Term{"text", "yellow"}, Term{"text", "has"}, Term{"text", "feelings"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim3)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query1 B C"].variableCache["*query1"] = []NodeValue{
		NodeValue{claim2},
		NodeValue{claim3},
	}
	expected_nodes["*query1 B C"].variableCache["B"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "yellow"}}},
		NodeValue{[]Term{Term{"text", "yellow"}}},
	}
	expected_nodes["*query1 B C"].variableCache["C"] = []NodeValue{
		NodeValue{[]Term{Term{"integer", "3"}}},
		NodeValue{[]Term{Term{"text", "feelings"}}},
	}
	expected_nodes["*query0 *query1 A B C"].variableCache["*query0"] = []NodeValue{
		NodeValue{claim},
		NodeValue{claim},
	}
	expected_nodes["*query0 *query1 A B C"].variableCache["*query1"] = []NodeValue{
		NodeValue{claim2},
		NodeValue{claim3},
	}
	expected_nodes["*query0 *query1 A B C"].variableCache["A"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "Sun"}}},
		NodeValue{[]Term{Term{"text", "Sun"}}},
	}
	expected_nodes["*query0 *query1 A B C"].variableCache["B"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "yellow"}}},
		NodeValue{[]Term{Term{"text", "yellow"}}},
	}
	expected_nodes["*query0 *query1 A B C"].variableCache["C"] = []NodeValue{
		NodeValue{[]Term{Term{"integer", "3"}}},
		NodeValue{[]Term{Term{"text", "feelings"}}},
	}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	claim4 := []Term{Term{"text", "Sky"}, Term{"text", "is"}, Term{"text", "blue"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim4)
	if updatedSubscriberOutput != false {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0 A B"].variableCache["*query0"] = []NodeValue{
		NodeValue{claim},
		NodeValue{claim4},
	}
	expected_nodes["*query0 A B"].variableCache["A"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "Sun"}}},
		NodeValue{[]Term{Term{"text", "Sky"}}},
	}
	expected_nodes["*query0 A B"].variableCache["B"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "yellow"}}},
		NodeValue{[]Term{Term{"text", "blue"}}},
	}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	retract1 := []Term{Term{"variable", ""}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription, updatedSubscriberOutput = subscriberRetractUpdate(subscription, retract1)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0 A B"].variableCache["*query0"] = []NodeValue{NodeValue{claim4}}
	expected_nodes["*query0 A B"].variableCache["A"] = []NodeValue{NodeValue{[]Term{Term{"text", "Sky"}}}}
	expected_nodes["*query0 A B"].variableCache["B"] = []NodeValue{NodeValue{[]Term{Term{"text", "blue"}}}}
	expected_nodes["*query0 *query1 A B C"] = makeNodeFromVariableNames([]string{"*query0", "*query1", "A", "B", "C"})
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))
}

func TestMakeSubscriberTwoPartsDifferentVariables(t *testing.T) {
	query_part1 := []Term{
		Term{"variable", ""},
		Term{"text", "is"},
		Term{"variable", "B"},
	}
	query_part2 := []Term{
		Term{"variable", ""},
		Term{"text", "has"},
		Term{"variable", "C"},
	}
	query := [][]Term{query_part1, query_part2}
	subscription := makeSubscriber(query)

	if len(subscription.nodes) != 3 {
		t.Error("Wrong number of nodes ", len(subscription.nodes))
	}
	expected_nodes := make(map[string]Node)
	expected_nodes["*query0 B"] = makeNodeFromVariableNames([]string{"*query0", "B"})
	expected_nodes["*query1 C"] = makeNodeFromVariableNames([]string{"*query1", "C"})
	expected_nodes["*query0 *query1 B C"] = makeNodeFromVariableNames([]string{"*query0", "*query1", "B", "C"})
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	if len(subscription.queryPartToUpdate) != len(query) {
		t.Error("length of queryPartToUpdate should match query len", len(subscription.queryPartToUpdate), len(query))
	}
	expected_queryPartToUpdate0 := []SubscriptionUpdateOptions{
		SubscriptionUpdateOptions{"*query1 C", "*query0 *query1 B C", []string{"*query0", "B"}}}
	expected_queryPartToUpdate1 := []SubscriptionUpdateOptions{
		SubscriptionUpdateOptions{"*query0 B", "*query0 *query1 B C", []string{"*query1", "C"}}}
	if !reflect.DeepEqual(subscription.queryPartToUpdate[0], expected_queryPartToUpdate0) {
		t.Error("subscription.queryPartToUpdate[0] does not match")
	}
	if !reflect.DeepEqual(subscription.queryPartToUpdate[1], expected_queryPartToUpdate1) {
		t.Error("subscription.queryPartToUpdate[1] does not match")
	}

	updatedSubscriberOutput := false

	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))
	claim := []Term{Term{"text", "Sun"}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim)
	if updatedSubscriberOutput != false {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0 B"].variableCache["*query0"] = []NodeValue{NodeValue{claim}}
	expected_nodes["*query0 B"].variableCache["B"] = []NodeValue{NodeValue{[]Term{Term{"text", "yellow"}}}}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))
	claim2 := []Term{Term{"text", "yellow"}, Term{"text", "has"}, Term{"integer", "3"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim2)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query1 C"].variableCache["*query1"] = []NodeValue{NodeValue{claim2}}
	expected_nodes["*query1 C"].variableCache["C"] = []NodeValue{NodeValue{[]Term{Term{"integer", "3"}}}}
	expected_nodes["*query0 *query1 B C"].variableCache["*query0"] = []NodeValue{NodeValue{claim}}
	expected_nodes["*query0 *query1 B C"].variableCache["*query1"] = []NodeValue{NodeValue{claim2}}
	expected_nodes["*query0 *query1 B C"].variableCache["B"] = []NodeValue{NodeValue{[]Term{Term{"text", "yellow"}}}}
	expected_nodes["*query0 *query1 B C"].variableCache["C"] = []NodeValue{NodeValue{[]Term{Term{"integer", "3"}}}}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}
	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	claim3 := []Term{Term{"text", "yellow"}, Term{"text", "has"}, Term{"text", "feelings"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim3)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query1 C"].variableCache["*query1"] = []NodeValue{
		NodeValue{claim2},
		NodeValue{claim3},
	}
	expected_nodes["*query1 C"].variableCache["C"] = []NodeValue{
		NodeValue{[]Term{Term{"integer", "3"}}},
		NodeValue{[]Term{Term{"text", "feelings"}}},
	}
	expected_nodes["*query0 *query1 B C"].variableCache["*query0"] = []NodeValue{
		NodeValue{claim},
		NodeValue{claim},
	}
	expected_nodes["*query0 *query1 B C"].variableCache["*query1"] = []NodeValue{
		NodeValue{claim2},
		NodeValue{claim3},
	}
	expected_nodes["*query0 *query1 B C"].variableCache["B"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "yellow"}}},
		NodeValue{[]Term{Term{"text", "yellow"}}},
	}
	expected_nodes["*query0 *query1 B C"].variableCache["C"] = []NodeValue{
		NodeValue{[]Term{Term{"integer", "3"}}},
		NodeValue{[]Term{Term{"text", "feelings"}}},
	}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	claim4 := []Term{Term{"text", "Sky"}, Term{"text", "is"}, Term{"text", "blue"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim4)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0 B"].variableCache["*query0"] = []NodeValue{
		NodeValue{claim},
		NodeValue{claim4},
	}
	expected_nodes["*query0 B"].variableCache["B"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "yellow"}}},
		NodeValue{[]Term{Term{"text", "blue"}}},
	}
	expected_nodes["*query0 *query1 B C"].variableCache["*query0"] = []NodeValue{
		NodeValue{claim},
		NodeValue{claim},
		NodeValue{claim4},
		NodeValue{claim4},
	}
	expected_nodes["*query0 *query1 B C"].variableCache["*query1"] = []NodeValue{
		NodeValue{claim2},
		NodeValue{claim3},
		NodeValue{claim2},
		NodeValue{claim3},
	}
	expected_nodes["*query0 *query1 B C"].variableCache["B"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "yellow"}}},
		NodeValue{[]Term{Term{"text", "yellow"}}},
		NodeValue{[]Term{Term{"text", "blue"}}},
		NodeValue{[]Term{Term{"text", "blue"}}},
	}
	expected_nodes["*query0 *query1 B C"].variableCache["C"] = []NodeValue{
		NodeValue{[]Term{Term{"integer", "3"}}},
		NodeValue{[]Term{Term{"text", "feelings"}}},
		NodeValue{[]Term{Term{"integer", "3"}}},
		NodeValue{[]Term{Term{"text", "feelings"}}},
	}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	retract1 := []Term{Term{"variable", ""}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription, updatedSubscriberOutput = subscriberRetractUpdate(subscription, retract1)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0 B"].variableCache["*query0"] = []NodeValue{NodeValue{claim4}}
	expected_nodes["*query0 B"].variableCache["B"] = []NodeValue{NodeValue{[]Term{Term{"text", "blue"}}}}
	expected_nodes["*query0 *query1 B C"].variableCache["*query0"] = []NodeValue{
		NodeValue{claim4},
		NodeValue{claim4},
	}
	expected_nodes["*query0 *query1 B C"].variableCache["*query1"] = []NodeValue{
		NodeValue{claim2},
		NodeValue{claim3},
	}
	expected_nodes["*query0 *query1 B C"].variableCache["B"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "blue"}}},
		NodeValue{[]Term{Term{"text", "blue"}}},
	}
	expected_nodes["*query0 *query1 B C"].variableCache["C"] = []NodeValue{
		NodeValue{[]Term{Term{"integer", "3"}}},
		NodeValue{[]Term{Term{"text", "feelings"}}},
	}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))
}

func TestMakeSubscriberOnePart(t *testing.T) {
	query_part1 := []Term{
		Term{"variable", "A"},
		Term{"text", "is"},
		Term{"variable", "B"},
	}
	query := [][]Term{query_part1}
	subscription := makeSubscriber(query)
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
	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	updatedSubscriberOutput := false

	claim := []Term{Term{"text", "Sun"}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0 A B"].variableCache["*query0"] = []NodeValue{NodeValue{claim}}
	expected_nodes["*query0 A B"].variableCache["A"] = []NodeValue{NodeValue{[]Term{Term{"text", "Sun"}}}}
	expected_nodes["*query0 A B"].variableCache["B"] = []NodeValue{NodeValue{[]Term{Term{"text", "yellow"}}}}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	// claim does not match
	claim2 := []Term{Term{"text", "Sun"}, Term{"text", "BAD"}, Term{"text", "yellow"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim2)
	if updatedSubscriberOutput != false {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	// $ is BAD does not match any previous claim, so this should do nothing
	retract1 := []Term{Term{"variable", ""}, Term{"text", "is"}, Term{"text", "BAD"}}
	subscription, updatedSubscriberOutput = subscriberRetractUpdate(subscription, retract1)
	if updatedSubscriberOutput != false {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	// $ is yellow matches a previous claim "Sun is yellow" so that claim should be removed
	retract2 := []Term{Term{"variable", ""}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription, updatedSubscriberOutput = subscriberRetractUpdate(subscription, retract2)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0 A B"] = makeNodeFromVariableNames([]string{"*query0", "A", "B"})
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))
}

func TestMakeSubscriberOnePartNoVariables(t *testing.T) {
	query_part1 := []Term{
		Term{"text", "Sun"},
		Term{"text", "is"},
		Term{"text", "yellow"},
	}
	query := [][]Term{query_part1}
	subscription := makeSubscriber(query)
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

	updatedSubscriberOutput := false

	claim := []Term{Term{"text", "Sun"}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0"].variableCache["*query0"] = []NodeValue{NodeValue{claim}}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	retract1 := []Term{Term{"postfix", ""}}
	subscription, updatedSubscriberOutput = subscriberRetractUpdate(subscription, retract1)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0"] = makeNodeFromVariableNames([]string{"*query0"})
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}
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
	subscription := makeSubscriber(query)
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
	// repr.Println(subscription, repr.Indent("  "), repr.OmitEmpty(true))

	updatedSubscriberOutput := false

	claim := []Term{Term{"text", "Sun"}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim)
	if updatedSubscriberOutput != false {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0"].variableCache["*query0"] = []NodeValue{NodeValue{claim}}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	claim2 := []Term{Term{"text", "Earth"}, Term{"text", "is"}, Term{"text", "green"}}
	subscription, updatedSubscriberOutput = subscriberClaimUpdate(subscription, claim2)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query1"].variableCache["*query1"] = []NodeValue{NodeValue{claim2}}
	expected_nodes["*query0 *query1"].variableCache["*query0"] = []NodeValue{NodeValue{claim}}
	expected_nodes["*query0 *query1"].variableCache["*query1"] = []NodeValue{NodeValue{claim2}}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	retract1 := []Term{Term{"variable", ""}, Term{"text", "is"}, Term{"text", "yellow"}}
	subscription, updatedSubscriberOutput = subscriberRetractUpdate(subscription, retract1)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0"] = makeNodeFromVariableNames([]string{"*query0"})
	expected_nodes["*query0 *query1"] = makeNodeFromVariableNames([]string{"*query0", "*query1"})
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	retract2 := []Term{Term{"text", "Earth"}, Term{"text", "is"}, Term{"text", "green"}}
	subscription, updatedSubscriberOutput = subscriberRetractUpdate(subscription, retract2)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0"] = makeNodeFromVariableNames([]string{"*query0"})
	expected_nodes["*query1"] = makeNodeFromVariableNames([]string{"*query1"})
	expected_nodes["*query0 *query1"] = makeNodeFromVariableNames([]string{"*query0", "*query1"})
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}
}

func TestSubscriberBatchSimple1(t *testing.T) {
	query_part1 := []Term{
		Term{"variable", "A"},
		Term{"text", "is"},
		Term{"variable", "B"},
	}
	query := [][]Term{query_part1}
	subscription := makeSubscriber(query)
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

	updatedSubscriberOutput := false

	batch_update := []BatchMessage{
		BatchMessage{"claim", [][]string{[]string{"text", "Sun"}, []string{"text", "is"}, []string{"text", "yellow"}}},
	}
	subscription, updatedSubscriberOutput = subscriberBatchUpdate(subscription, batch_update)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	claim := []Term{Term{"text", "Sun"}, Term{"text", "is"}, Term{"text", "yellow"}}
	expected_nodes["*query0 A B"].variableCache["*query0"] = []NodeValue{NodeValue{claim}}
	expected_nodes["*query0 A B"].variableCache["A"] = []NodeValue{NodeValue{[]Term{Term{"text", "Sun"}}}}
	expected_nodes["*query0 A B"].variableCache["B"] = []NodeValue{NodeValue{[]Term{Term{"text", "yellow"}}}}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	// 1. claim does not match
	// 2. $ is BAD does not match any previous claim, so this should do nothing
	batch_update2 := []BatchMessage{
		BatchMessage{"claim", [][]string{[]string{"text", "Sun"}, []string{"text", "BAD"}, []string{"text", "yellow"}}},
		BatchMessage{"retract", [][]string{[]string{"variable", ""}, []string{"text", "is"}, []string{"text", "BAD"}}},
	}
	subscription, updatedSubscriberOutput = subscriberBatchUpdate(subscription, batch_update2)
	if updatedSubscriberOutput != false {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	// $ is yellow matches a previous claim "Sun is yellow" so that claim should be removed
	batch_update3 := []BatchMessage{
		BatchMessage{"retract", [][]string{[]string{"variable", ""}, []string{"text", "is"}, []string{"text", "yellow"}}},
	}
	subscription, updatedSubscriberOutput = subscriberBatchUpdate(subscription, batch_update3)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}
	expected_nodes["*query0 A B"] = makeNodeFromVariableNames([]string{"*query0", "A", "B"})
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}
}

func TestSubscriberBatchBigClaimAndRetract(t *testing.T) {
	query_part1 := []Term{
		Term{"variable", "A"},
		Term{"text", "is"},
		Term{"variable", "B"},
	}
	query := [][]Term{query_part1}
	subscription := makeSubscriber(query)
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

	updatedSubscriberOutput := false

	batch_update := []BatchMessage{
		BatchMessage{"claim", [][]string{[]string{"text", "Sun"}, []string{"text", "is"}, []string{"text", "yellow"}}},
		BatchMessage{"claim", [][]string{[]string{"text", "Sky"}, []string{"text", "is"}, []string{"text", "blue"}}},
		BatchMessage{"claim", [][]string{[]string{"text", "Sky"}, []string{"text", "is"}, []string{"text", "black"}}},
		BatchMessage{"claim", [][]string{[]string{"text", "Moon"}, []string{"text", "is"}, []string{"text", "white"}}},
		BatchMessage{"claim", [][]string{[]string{"text", "Moon"}, []string{"text", "BAD"}, []string{"text", "black"}}},
		BatchMessage{"claim", [][]string{[]string{"text", "Moon"}, []string{"text", "is"}, []string{"text", "grey"}, []string{"text", "colored"}}},
	}
	subscription, updatedSubscriberOutput = subscriberBatchUpdate(subscription, batch_update)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}

	expected_nodes["*query0 A B"].variableCache["*query0"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "Sun"}, Term{"text", "is"}, Term{"text", "yellow"}}},
		NodeValue{[]Term{Term{"text", "Sky"}, Term{"text", "is"}, Term{"text", "blue"}}},
		NodeValue{[]Term{Term{"text", "Sky"}, Term{"text", "is"}, Term{"text", "black"}}},
		NodeValue{[]Term{Term{"text", "Moon"}, Term{"text", "is"}, Term{"text", "white"}}},
	}
	expected_nodes["*query0 A B"].variableCache["A"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "Sun"}}},
		NodeValue{[]Term{Term{"text", "Sky"}}},
		NodeValue{[]Term{Term{"text", "Sky"}}},
		NodeValue{[]Term{Term{"text", "Moon"}}},
	}
	expected_nodes["*query0 A B"].variableCache["B"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "yellow"}}},
		NodeValue{[]Term{Term{"text", "blue"}}},
		NodeValue{[]Term{Term{"text", "black"}}},
		NodeValue{[]Term{Term{"text", "white"}}},
	}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	batch_update2 := []BatchMessage{
		BatchMessage{"retract", [][]string{[]string{"text", "Sky"}, []string{"text", "is"}, []string{"text", "black"}}},
	}
	subscription, updatedSubscriberOutput = subscriberBatchUpdate(subscription, batch_update2)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}
	expected_nodes["*query0 A B"].variableCache["*query0"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "Sun"}, Term{"text", "is"}, Term{"text", "yellow"}}},
		NodeValue{[]Term{Term{"text", "Sky"}, Term{"text", "is"}, Term{"text", "blue"}}},
		NodeValue{[]Term{Term{"text", "Moon"}, Term{"text", "is"}, Term{"text", "white"}}},
	}
	expected_nodes["*query0 A B"].variableCache["A"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "Sun"}}},
		NodeValue{[]Term{Term{"text", "Sky"}}},
		NodeValue{[]Term{Term{"text", "Moon"}}},
	}
	expected_nodes["*query0 A B"].variableCache["B"] = []NodeValue{
		NodeValue{[]Term{Term{"text", "yellow"}}},
		NodeValue{[]Term{Term{"text", "blue"}}},
		NodeValue{[]Term{Term{"text", "white"}}},
	}
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}

	batch_update3 := []BatchMessage{
		BatchMessage{"retract", [][]string{[]string{"postfix", "X"}}},
	}
	subscription, updatedSubscriberOutput = subscriberBatchUpdate(subscription, batch_update3)
	if updatedSubscriberOutput != true {
		t.Error("Flag for updated subscriber output is wrong!", updatedSubscriberOutput)
	}
	expected_nodes["*query0 A B"] = makeNodeFromVariableNames([]string{"*query0", "A", "B"})
	if !reflect.DeepEqual(subscription.nodes, expected_nodes) {
		t.Error("Contents of nodes is wrong", subscription.nodes)
	}
}
