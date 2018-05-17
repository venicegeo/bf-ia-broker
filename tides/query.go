package tides

import (
	"fmt"

	"github.com/venicegeo/bf-ia-broker/model"
	"github.com/venicegeo/bf-ia-broker/util"
)

var httpRequestKnownJSONWithObject = util.ReqByObjJSON

func QueryMultipleTides(context *Context, input Input) (*Output, error) {
	var out Output

	util.LogAudit(context, util.LogAuditInput{
		Actor: "anon user", Action: "POST", Actee: context.TidesURL, Message: "Requesting tide information", Severity: util.INFO,
	})
	if _, err := httpRequestKnownJSONWithObject("POST", context.TidesURL, "", input, &out); err != nil {
		return nil, err
	}
	util.LogAudit(context, util.LogAuditInput{
		Actor: context.TidesURL, Action: "POST response", Actee: "anon user", Message: "Retrieving tide information", Severity: util.INFO,
	})

	return &out, nil
}

// AddTidesToSearchResults does an *in-place* modification of the input broker
// search results to augment them with tides data
func AddTidesToSearchResults(context *Context, results []model.BrokerSearchResult) error {
	basicResults := make([]model.BasicBrokerResult, len(results))
	for i, result := range results {
		basicResults[i] = result.BasicBrokerResult
	}

	input, err := InputForBasicBrokerResults(basicResults)
	if err != nil {
		return err
	}

	output, err := QueryMultipleTides(context, *input)
	if err != nil {
		return err
	}

	tidesDataArr := OutputToTidesData(*output)
	if len(tidesDataArr) != len(results) {
		return fmt.Errorf("Length mismatch between tides output and input data;\ninput(len:%d)=%v\noutput(len:%d)=%v",
			len(input.Locations), input, len(output.Locations), output,
		)
	}

	for i := range results {
		results[i].TidesData = &tidesDataArr[i]
	}

	return nil
}

func GetSingleTidesData(context *Context, target model.BasicBrokerResult) (*model.TidesData, error) {
	return nil, nil
}
