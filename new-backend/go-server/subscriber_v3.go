package main

import (
	"go.uber.org/zap"
)

type Subscription3 struct {
	query                   [][]Term
	/* Fact table cache = queryPartMatchingFacts
	   For each query part, a fact table filtered to only facts matching that query part
	   Index of "queryPartMatchingFacts" corresponds to index in "query"
	 */
	queryPartMatchingFacts  []map[string]Fact
}

func setupSubscriber(query [][]Term, preExistingFacts map[string]Fact) Subscription3 {
	subscriber := Subscription3{query, make([]map[string]Fact, len(query))}
	// Initialize the queryPartMatchingFacts cache
	for i, queryPart := range query {
		subscriber.queryPartMatchingFacts[i] = make(map[string]Fact)
		for factKey, fact := range preExistingFacts {
			empty_env := QueryResult{map[string]Term{}}
			did_match, _ := fact_match(Fact{queryPart}, fact, empty_env)
			if did_match {
				subscriber.queryPartMatchingFacts[i][factKey] = fact
			}
		}
	}
	return subscriber
}

func updateQueryPartMatchingFactsFromRetract(sub *Subscription3, factQuery Fact) bool {
	anythingChanged := false
	for i := 0; i < len((*sub).queryPartMatchingFacts); i++ {
		prevSize := len((*sub).queryPartMatchingFacts[i])
		retract(&(*sub).queryPartMatchingFacts[i], factQuery) // can we modify the subscriber cache in place like this?
		if prevSize != len((*sub).queryPartMatchingFacts[i]) {
			anythingChanged = true
		}
	}
	return anythingChanged
}

func updateQueryPartMatchingFactsFromClaim(sub *Subscription3, fact Fact) bool {
	anythingChanged := false
	for i := 0; i < len((*sub).queryPartMatchingFacts); i++ {
		did_match, _ := fact_match(Fact{(*sub).query[i]}, fact, QueryResult{})
		if did_match {
			claim(&(*sub).queryPartMatchingFacts[i], fact) // can we modify the subscriber cache in place like this?
			anythingChanged = true
		}
	}
	return anythingChanged
}

func subscriberBatchUpdateV3(sub *Subscription3, batch_messages []BatchMessage) bool {
	updatedSubscriberOutput := false
	for _, batch_message := range batch_messages {
		terms := make([]Term, len(batch_message.Fact))
		for j, term := range batch_message.Fact {
			terms[j] = Term{term[0], term[1]}
		}
		if batch_message.Type == "claim" {
			anythingChanged := updateQueryPartMatchingFactsFromClaim(sub, Fact{terms})
			updatedSubscriberOutput = updatedSubscriberOutput || anythingChanged
		} else if batch_message.Type == "retract" {
			anythingChanged := updateQueryPartMatchingFactsFromRetract(sub, Fact{terms})
			updatedSubscriberOutput = updatedSubscriberOutput || anythingChanged
		} else if batch_message.Type == "death" {
			// TODO: don't reply logic from server.go that also does this retract
			dying_source := batch_message.Fact[0][1]
			clearSourceClaims := []Term{Term{"id", dying_source}, Term{"postfix", ""}}
			// TODO: clearSourceSubscriptions := []Term{Term{"text", "subscription"}, Term{"id", dying_source}, Term{"postfix", ""}}
			anythingChanged := updateQueryPartMatchingFactsFromRetract(sub, Fact{clearSourceClaims})
			updatedSubscriberOutput = updatedSubscriberOutput || anythingChanged
		}
	}
	return updatedSubscriberOutput
}

func subscriberCollectSolutions(sub Subscription3) []QueryResult {
	subQueryAsFact := make([]Fact, len(sub.query))
	for i, val := range sub.query {
		subQueryAsFact[i] = Fact{val}
	}
	return collect_solutions_v3(sub.queryPartMatchingFacts, subQueryAsFact, 0, QueryResult{})
}

func startSubscriberV3(subscriptionData Subscription, notifications chan<- Notification, preExistingFacts map[string]Fact) {
	subscriber := setupSubscriber(subscriptionData.Query, preExistingFacts)
	zap.L().Info("inside startSubscriber v3")
	for batch_messages := range subscriptionData.batch_messages {
		updatedResults := subscriberBatchUpdateV3(&subscriber, batch_messages)
		if updatedResults {
			results := subscriberCollectSolutions(subscriber)
			// TODO: sort results?
			results_as_str := marshal_query_result(results)
			notifications <- Notification{subscriptionData.Source, subscriptionData.Id, results_as_str}
		}
	}
	subscriptionData.dead.Done()
}