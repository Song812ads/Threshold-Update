package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/song/service"
)

type Threshold struct {
	start  float32
	end    float32
	status string
}

func mapThresholdsToString(thresholds []Threshold) string {
	var builder strings.Builder

	for _, th := range thresholds {
		builder.WriteString(fmt.Sprintf(" WHEN dat > %.2f AND dat < %.2f THEN  \"%s\"  ", th.start, th.end, th.status))
	}

	return builder.String()
}

func UpdateThreshold(factory_id string, field string, thres []Threshold, method string) error {

	c := http.Client{Timeout: time.Duration(1) * time.Second}

	config_name := fmt.Sprintf("%s_%s_Config", factory_id, field)

	stream_name := fmt.Sprintf("%s_%s_Stream", factory_id, field)

	rule_name := fmt.Sprintf("%s_%s_Rule", factory_id, field)

	_ = rule_name

	url_config := fmt.Sprintf("http://localhost:59720/metadata/sources/edgex/confKeys/%s", config_name)

	url_stream := "http://localhost:59720/streams"

	url_rules := "http://localhost:59720/rules"

	device := service.MapForFactory(factory_id)

	data := map[string]interface{}{
		fmt.Sprintf("%s", config_name): map[string]interface{}{
			"topic":       fmt.Sprintf("edgex/events/device/%s/#/#", device),
			"messageType": "request",
		},
	}

	jsonConfig, err := json.MarshalIndent(data, "", "  ")

	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return err
	}
	if _, err := service.SendPutRequest(url_config, jsonConfig); err != nil {
		fmt.Println("Error Post Config request: ", err)
		return err
	}

	if method == "POST" {
		data := map[string]interface{}{
			"sql": fmt.Sprintf("CREATE STREAM %s() WITH (TYPE=\"edgex\", FORMAT = \"JSON\", SHARED = \"true\", CONF_KEY = \"%s\")", stream_name, config_name),
		}

		jsonStream, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Println("Error create Stream:", err)
			return err
		}

		if err := service.PostRequest(c, url_stream, jsonStream); err != nil {
			fmt.Println("Post stream fail: ", err)
			return err
		}
	}

	query_String := mapThresholdsToString(thres)

	data = map[string]interface{}{
		"id":  rule_name,
		"sql": fmt.Sprintf(" SELECT %s.`d-ata` AS dat, CASE %s ELSE \"M0\" END AS dataClass FROM %s  WHERE  %s.`d-ata` != NIL", stream_name, query_String, stream_name, stream_name),
		"actions": []map[string]interface{}{
			{
				"mqtt": map[string]interface{}{
					"server":       "tcp://172.18.0.21:1883",
					"topic":        "demoSink",
					"sendSingle":   true,
					"dataTemplate": fmt.Sprintf("{\"FactoryId\":\"%s\", \"Name\": \"%s\", \"Level\": \"{{.dataClass}}\", \"Value\": {{.dat}} }", factory_id, field),
				},
			},
		},
	}

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.Encode(data)
	fmt.Println(buf.String())

	if method == "POST" {
		if err := service.PostRequest(c, url_rules, buf.Bytes()); err != nil {
			fmt.Println("Post or Update Rule fail: ", err)
			return err
		}
	} else if method == "PUT" {
		if _, err := service.SendPutRequest(url_rules, buf.Bytes()); err != nil {
			fmt.Println("Post or Update Rule fail: ", err)
			return err
		}
	}

	return nil
}

func main() {

	thresholds := []Threshold{
		{start: 0, end: 100, status: "Low"},
		{start: 101, end: 200, status: "Medium"},
		{start: 201, end: 300, status: "High"},
	}
	if err := UpdateThreshold("hihi1", "haha1", thresholds, "POST"); err != nil {
		fmt.Println(err)
	}
}
