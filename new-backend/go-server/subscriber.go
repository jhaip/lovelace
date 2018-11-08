package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

/*
A subscriber is a goroutine that receives every claim and retract
and sends notifications
*/

type NodeValue struct {
	terms []Term
}

type Node struct {
	variableCache map[string][]NodeValue
}

type SubscriptionUpdateOptions struct {
	sourceNodeKey  string
	destNodeKey    string
	variablesToAdd []string
}

type Subscription2 struct {
	query                   [][]Term
	queryPartToUpdate       map[int][]SubscriptionUpdateOptions
	nodes                   map[string]Node
	outputVariables         []string
	outputVariablesNodesKey string
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
			addedVariables = append(addedVariables, y)
		}
	}
	if len(res) == len(a) {
		return res, addedVariables, false
	}
	sort.Strings(res)
	if len(a) == 1 && strings.HasPrefix(a[0], "*query") {
		return res, addedVariables, true
	}
	if len(b) == 1 && strings.HasPrefix(b[0], "*query") {
		return res, addedVariables, true
	}
	return res, addedVariables, atLeastOneOverlap
}

func makeNodeFromVariableNames(variableNames []string) Node {
	variableCache := make(map[string][]NodeValue)
	for _, variableName := range variableNames {
		variableCache[variableName] = make([]NodeValue, 0)
	}
	return Node{variableCache}
}

func getVariableNamesFromNode(node Node) []string {
	keys := make([]string, len(node.variableCache))
	i := 0
	for k := range node.variableCache {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func makeSubscriber(source string, id string, query [][]Term) Subscription2 {
	subscriber := Subscription2{query, make(map[int][]SubscriptionUpdateOptions), make(map[string]Node), make([]string, 0), ""}
	originalSubscriberNodeKeys := make([]string, 0)
	outputVariablesMap := make(map[string]bool)
	for i, queryPart := range query {
		queryPartVariableNames := append([]string{"*query" + strconv.Itoa(i)}, getVariableTermNames(queryPart)...)
		for _, queryPartVariableName := range queryPartVariableNames {
			outputVariablesMap[queryPartVariableName] = true
		}
		variableTermKey := getVariableTermKey(queryPartVariableNames)
		subscriber.nodes[variableTermKey] = makeNodeFromVariableNames(queryPartVariableNames)
		originalSubscriberNodeKeys = append(originalSubscriberNodeKeys, variableTermKey)
		subscriber.queryPartToUpdate[i] = make([]SubscriptionUpdateOptions, 0)
	}
	for _, originalSubscriberNodeKey := range originalSubscriberNodeKeys {
		for i, originalSubscriberNodeKey2 := range originalSubscriberNodeKeys {
			combinedKeys, addedVariables, matched := getCombinedKeyNames(
				getVariableNamesFromNode(subscriber.nodes[originalSubscriberNodeKey]),
				getVariableNamesFromNode(subscriber.nodes[originalSubscriberNodeKey2]))
			if matched {
				variableTermKey := getVariableTermKey(combinedKeys)
				subscriber.nodes[variableTermKey] = makeNodeFromVariableNames(combinedKeys)
				subscriber.queryPartToUpdate[i] = append(subscriber.queryPartToUpdate[i],
					SubscriptionUpdateOptions{
						originalSubscriberNodeKey,
						variableTermKey,
						addedVariables})
			}
		}
	}
	for outputVariable, _ := range outputVariablesMap {
		subscriber.outputVariables = append(subscriber.outputVariables, outputVariable)
	}
	sort.Strings(subscriber.outputVariables)
	subscriber.outputVariablesNodesKey = getVariableTermKey(subscriber.outputVariables)
	return subscriber
}

func populateFirstLayerFromMatchResults(queryPartIndex int, matchResults QueryResult, sub Subscription2, claim []Term) (Subscription2, bool) {
	fmt.Println("MATCH RESULTS:")
	fmt.Println(matchResults)
	fmt.Println(matchResults.Result)
	fmt.Println("--")
	matchResultVariableNames := make([]string, 0)
	for variableName, _ := range matchResults.Result {
		matchResultVariableNames = append(matchResultVariableNames, variableName)
	}
	fmt.Println("----")
	querySourceVariableName := "*query" + strconv.Itoa(queryPartIndex)
	queryPartVariableNames := append(
		[]string{querySourceVariableName},
		matchResultVariableNames...)
	sort.Strings(queryPartVariableNames)
	variableTermKey := getVariableTermKey(queryPartVariableNames)
	variableCache := make(map[string][]NodeValue)
	for variableName, matchResultTerm := range matchResults.Result {
		variableCache[variableName] = append(
			sub.nodes[variableTermKey].variableCache[variableName],
			NodeValue{[]Term{matchResultTerm}},
		)
	}
	variableCache[querySourceVariableName] = append(
		sub.nodes[variableTermKey].variableCache[querySourceVariableName],
		NodeValue{claim},
	)
	sub.nodes[variableTermKey] = Node{variableCache}
	updatedSubscriberOutput := (variableTermKey == sub.outputVariablesNodesKey)
	return sub, updatedSubscriberOutput
}

func copyNode(node Node) Node {
	variableCache := make(map[string][]NodeValue)
	for k, v := range node.variableCache {
		nodeValues := make([]NodeValue, len(v))
		for i, nodeValue := range v {
			nodeValues[i] = nodeValue
		}
		variableCache[k] = nodeValues
	}
	return Node{variableCache}
}

func getLengthOfNodeVariableCache(node Node) int {
	for _, sourceVariableCache := range node.variableCache {
		return len(sourceVariableCache)
	}
	return 0
}

func addQueryResultToWholeVariableCache(queryPartIndex int, subscriptionUpdateOptions SubscriptionUpdateOptions, matchResults QueryResult, sub Subscription2, claim []Term) (Subscription2, bool) {
	newSourceNode := copyNode(sub.nodes[subscriptionUpdateOptions.sourceNodeKey])
	thingsToAddToDestinationNode := make(map[string][]NodeValue)

	for destVariableName, _ := range sub.nodes[subscriptionUpdateOptions.destNodeKey].variableCache {
		thingsToAddToDestinationNode[destVariableName] = make([]NodeValue, 0)
	}

	lengthOfSourceCache := getLengthOfNodeVariableCache(newSourceNode)

	for i := 0; i < lengthOfSourceCache; i++ {
		elementAtOffsetHasNoOverlapOrMatchingOverlap := true
		for sourceVariableName, sourceVariableCache := range newSourceNode.variableCache {
			_, matchResultsHasSourceVariable := matchResults.Result[sourceVariableName]
			if matchResultsHasSourceVariable {
				if matchResults.Result[sourceVariableName].Type != sourceVariableCache[i].terms[0].Type ||
					matchResults.Result[sourceVariableName].Value != sourceVariableCache[i].terms[0].Value {
					elementAtOffsetHasNoOverlapOrMatchingOverlap = false
					break
				}
			}
		}
		if elementAtOffsetHasNoOverlapOrMatchingOverlap {
			// Build the new thing to copy to the destination node
			// 1. Copy the variables from the source node
			for sourceVariableName, sourceVariableCache := range newSourceNode.variableCache {
				thingsToAddToDestinationNode[sourceVariableName] = append(
					thingsToAddToDestinationNode[sourceVariableName],
					sourceVariableCache[i],
				)
			}
			// 2. Add in the new variables from the matchResult
			for _, variableName := range subscriptionUpdateOptions.variablesToAdd {
				if strings.HasPrefix(variableName, "*query") {
					thingsToAddToDestinationNode[variableName] = append(
						thingsToAddToDestinationNode[variableName],
						NodeValue{claim},
					)
				} else {
					thingsToAddToDestinationNode[variableName] = append(
						thingsToAddToDestinationNode[variableName],
						NodeValue{[]Term{matchResults.Result[variableName]}},
					)
				}
			}
		} else {
			fmt.Println("elementAtOffsetHasNoOverlapOrMatchingOverlap is FALSE!")
		}
	}

	newDestNode := copyNode(sub.nodes[subscriptionUpdateOptions.destNodeKey])
	thingsToAddToDestinationNodeWasntEmpty := false
	for variableName, nodeValues := range thingsToAddToDestinationNode {
		if len(nodeValues) > 0 {
			thingsToAddToDestinationNodeWasntEmpty = true
		}
		newDestNode.variableCache[variableName] = append(newDestNode.variableCache[variableName], nodeValues...)
	}
	sub.nodes[subscriptionUpdateOptions.destNodeKey] = newDestNode
	updatedSubscriberOutput := (subscriptionUpdateOptions.destNodeKey == sub.outputVariablesNodesKey && thingsToAddToDestinationNodeWasntEmpty)
	return sub, updatedSubscriberOutput
}

func subscriberClaimUpdate(sub Subscription2, claim []Term) (Subscription2, bool) {
	updatedSubscriberOutput := false
	for i, query_part := range sub.query {
		queryPartUpdatedSubscriberOutput := false
		match, matchResults := fact_match(Fact{query_part}, Fact{claim}, QueryResult{})
		if match {
			sub, queryPartUpdatedSubscriberOutput = populateFirstLayerFromMatchResults(i, matchResults, sub, claim)
			if queryPartUpdatedSubscriberOutput {
				updatedSubscriberOutput = true
			}
			for _, subscriptionUpdateOptions := range sub.queryPartToUpdate[i] {
				sub, queryPartUpdatedSubscriberOutput = addQueryResultToWholeVariableCache(i, subscriptionUpdateOptions, matchResults, sub, claim)
				if queryPartUpdatedSubscriberOutput {
					updatedSubscriberOutput = true
				}
			}
		}
	}
	return sub, updatedSubscriberOutput
}

func subscriberRetractUpdate(sub Subscription2, query []Term) (Subscription2, bool) {
	updatedSubscriberOutput := false
	for nodeKey, node := range sub.nodes {
		lengthOfNodeCache := getLengthOfNodeVariableCache(node)
		updatedNode := Node{make(map[string][]NodeValue)}
		for variableName, _ := range node.variableCache {
			updatedNode.variableCache[variableName] = make([]NodeValue, 0)
		}
		for i := 0; i < lengthOfNodeCache; i++ {
			cacheRowIsOk := true
			for variableName, variableCache := range node.variableCache {
				if strings.HasPrefix(variableName, "*query") {
					match, _ := fact_match(Fact{query}, Fact{variableCache[i].terms}, QueryResult{})
					if match {
						cacheRowIsOk = false
						updatedSubscriberOutput = true
						break
					}
				}
			}
			if cacheRowIsOk {
				for variableName, variableCache := range node.variableCache {
					updatedNode.variableCache[variableName] = append(
						updatedNode.variableCache[variableName],
						variableCache[i],
					)
				}
			}
		}

		sub.nodes[nodeKey] = updatedNode
	}
	return sub, updatedSubscriberOutput
}

func subscriberBatchUpdate(sub Subscription2, batch_messages []BatchMessage) (Subscription2, bool) {
	updatedSubscriberOutput := false
	for _, batch_message := range batch_messages {
		terms := make([]Term, len(batch_message.Fact))
		for j, term := range batch_message.Fact {
			terms[j] = Term{term[0], term[1]}
		}
		batchMessageUpdatedSubscriberOutput := false
		if batch_message.Type == "claim" {
			sub, batchMessageUpdatedSubscriberOutput = subscriberClaimUpdate(sub, terms)
		} else if batch_message.Type == "retract" {
			sub, batchMessageUpdatedSubscriberOutput = subscriberRetractUpdate(sub, terms)
		}
		if batchMessageUpdatedSubscriberOutput {
			updatedSubscriberOutput = true
		}
	}
	return sub, updatedSubscriberOutput
}

func subscriber(batch_messages <-chan string) {

}
