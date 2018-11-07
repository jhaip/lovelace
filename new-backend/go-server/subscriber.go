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
	terms   []Term
	sources []string
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
	query             [][]Term
	queryPartToUpdate map[int][]SubscriptionUpdateOptions
	nodes             map[string]Node
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
	subscriber := Subscription2{query, make(map[int][]SubscriptionUpdateOptions), make(map[string]Node)}
	originalSubscriberNodeKeys := make([]string, 0)
	for i, queryPart := range query {
		queryPartVariableNames := append([]string{"*query" + strconv.Itoa(i)}, getVariableTermNames(queryPart)...)
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
	return subscriber
}

func populateFirstLayerFromMatchResults(queryPartIndex int, matchResults QueryResult, sub Subscription2, claim []Term) Subscription2 {
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
			NodeValue{[]Term{matchResultTerm}, []string{strconv.Itoa(queryPartIndex)}},
		)
	}
	variableCache[querySourceVariableName] = append(
		sub.nodes[variableTermKey].variableCache[querySourceVariableName],
		NodeValue{claim, []string{strconv.Itoa(queryPartIndex)}},
	)
	sub.nodes[variableTermKey] = Node{variableCache}
	return sub
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

func addQueryResultToWholeVariableCache(queryPartIndex int, subscriptionUpdateOptions SubscriptionUpdateOptions, matchResults QueryResult, sub Subscription2, claim []Term) Subscription2 {
	newSourceNode := copyNode(sub.nodes[subscriptionUpdateOptions.sourceNodeKey])
	for _, variableName := range subscriptionUpdateOptions.variablesToAdd {
		for _, oldCacheNodeValues := range newSourceNode.variableCache {
			newSourceNode.variableCache[variableName] = make([]NodeValue, len(oldCacheNodeValues))
			for i, oldCacheNodeValue := range oldCacheNodeValues {
				if strings.HasPrefix(variableName, "*query") {
					newSourceNode.variableCache[variableName][i] = NodeValue{
						claim,
						append(oldCacheNodeValue.sources, strconv.Itoa(queryPartIndex)),
					}
				} else {
					newSourceNode.variableCache[variableName][i] = NodeValue{
						[]Term{matchResults.Result[variableName]},
						append(oldCacheNodeValue.sources, strconv.Itoa(queryPartIndex)),
					}
				}
			}
			// Intentionally break.  The for loop was only used to get the length of any arbitary map []NodeValue
			break
		}
	}
	newDestNode := copyNode(sub.nodes[subscriptionUpdateOptions.destNodeKey])
	for variableName, nodeValues := range newSourceNode.variableCache {
		newDestNode.variableCache[variableName] = append(newDestNode.variableCache[variableName], nodeValues...)
	}
	sub.nodes[subscriptionUpdateOptions.destNodeKey] = newDestNode
	return sub
}

func subscriberClaimUpdate(sub Subscription2, claim []Term) Subscription2 {
	for i, query_part := range sub.query {
		match, matchResults := fact_match(Fact{query_part}, Fact{claim}, QueryResult{})
		if match {
			sub = populateFirstLayerFromMatchResults(i, matchResults, sub, claim)
			for _, subscriptionUpdateOptions := range sub.queryPartToUpdate[i] {
				sub = addQueryResultToWholeVariableCache(i, subscriptionUpdateOptions, matchResults, sub, claim)
			}
		}
	}
	return sub
}

func subscriber(batch_messages <-chan string) {

}
