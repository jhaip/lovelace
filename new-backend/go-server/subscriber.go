package main

import (
	"sort"
	"strconv"
	"strings"
)

/*
A subscriber is a goroutine that receives every claim and retract
and sends notifications
*/

type NodeValue struct {
	terms   []Term
	sources []string
}

type Node struct {
	variables   []string
	resultCache []NodeValue
}

type SubscriptionUpdateOptions struct {
	sourceNodeKey  string
	destNodeKey    string
	variablesToAdd []string
}

type Subscription2 struct {
	queryPartToUpdate map[int][]SubscriptionUpdateOptions
	// nodes             []Node
	nodes map[string]Node
}

func getVariableTermNames(terms []Term) []string {
	variableTermNames := make([]string, 0)

	for _, term := range terms {
		if (term.Type == "variable" || term.Type == "postfix") && term.Value != "" {
			variableTermNames = append(variableTermNames, term.Value)
		}
	}
	sort.Strings(variableTermNames)
	return variableTermNames
}

func getVariableTermKey(termNames []string) string {
	return strings.Join(termNames, " ")
}

func getCombinedKeyNames(a []string, b []string) ([]string, []string, bool) {
	// given A,B,C and update C,D
	// output A,B,C,D and update D
	res := make([]string, 0)
	addedVariables := make([]string, 0)
	keys := make(map[string]bool)
	atLeastOneOverlap := false
	for _, x := range a {
		res = append(res, x)
		keys[x] = true
	}
	for _, y := range b {
		if keys[y] {
			atLeastOneOverlap = true
		} else {
			res = append(res, y)
			addedVariables = append(res, y)
		}
	}
	if len(res) == len(a) {
		return res, addedVariables, false
	}
	sort.Strings(res)
	if len(a) == 1 && strings.HasPrefix(a[0], "__") {
		return res, addedVariables, true
	}
	if len(b) == 1 && strings.HasPrefix(b[0], "__") {
		return res, addedVariables, true
	}
	return res, addedVariables, atLeastOneOverlap
}

func makeSubscriber(source string, id string, query [][]Term) Subscription2 {
	subscriber := Subscription2{make(map[int][]SubscriptionUpdateOptions), make(map[string]Node)}
	originalSubscriberNodeKeys := make([]string, 0)
	for i, queryPart := range query {
		queryPartVariableNames := append([]string{"__" + strconv.Itoa(i)}, getVariableTermNames(queryPart)...)
		variableTermKey := getVariableTermKey(queryPartVariableNames)
		subscriber.nodes[variableTermKey] = Node{queryPartVariableNames, make([]NodeValue, 0)}
		originalSubscriberNodeKeys = append(originalSubscriberNodeKeys, variableTermKey)
		subscriber.queryPartToUpdate[i] = make([]SubscriptionUpdateOptions, 0)
	}
	for _, originalSubscriberNodeKey := range originalSubscriberNodeKeys {
		for i, originalSubscriberNodeKey2 := range originalSubscriberNodeKeys {
			combinedKeys, addedVariables, matched := getCombinedKeyNames(
				subscriber.nodes[originalSubscriberNodeKey].variables,
				subscriber.nodes[originalSubscriberNodeKey2].variables)
			if matched {
				variableTermKey := getVariableTermKey(combinedKeys)
				subscriber.nodes[variableTermKey] = Node{combinedKeys, make([]NodeValue, 0)}
				subscriber.queryPartToUpdate[i] = append(subscriber.queryPartToUpdate[i],
					SubscriptionUpdateOptions{
						originalSubscriberNodeKey,
						originalSubscriberNodeKey2,
						addedVariables})
			}
		}
	}
	return subscriber
}

func subscriber(batch_messages <-chan string) {

}
