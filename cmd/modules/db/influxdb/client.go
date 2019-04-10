package influxdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mirrors_status/cmd/log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

type Client struct {
	UseHttps bool
	Username string
	Password string
	Host     string
	Port     int
	DbName   string

	client *http.Client
}

func (c *Client) Exec(query string) (data Data, err error) {
	var protocol string
	if c.UseHttps {
		protocol = "https://"
	} else {
		protocol = "http://"
	}
	api := fmt.Sprintf("%s%s:%d/query", protocol, c.Host, c.Port)

	values := url.Values{}
	values.Set("db", c.DbName)
	values.Add("q", query)

	req, err := http.NewRequest("GET", api+"?"+values.Encode(), nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(c.Username, c.Password)
	if c.client == nil {
		c.client = &http.Client{}
	}
	res, err := c.client.Do(req)
	if err != nil {
		return
	}
	if res.StatusCode >= 300 || res.StatusCode < 200 {
		res, _ := ioutil.ReadAll(res.Body)
		err = fmt.Errorf("influxdb server return err: %s", string(res))
		return
	}
	json.NewDecoder(res.Body).Decode(&data)
	return
}

type Value struct {
	Value     interface{}
	Timestamp time.Time
}

func (c *Client) Select(measurement string, tags []string, where map[string]interface{}) (
	data [][]interface{}, err error) {
	clause := make([]string, 0)
	for tag, val := range where {
		typ := reflect.TypeOf(val)
		switch typ.Kind() {
		case reflect.Int, reflect.Int64:
			clause = append(clause, fmt.Sprintf(`"%s" = %d`, tag, val))
		case reflect.String:
			clause = append(clause, fmt.Sprintf(`"%s" = '%s'`, tag, val))
		}
	}
	query := fmt.Sprintf(`select %s from %s`, strings.Join(tags, ","), measurement)
	if len(clause) != 0 {
		query += " where " + strings.Join(clause, " and ")
	}
	rawdata, err := c.Exec(query)
	if err != nil {
		return
	}
	if len(rawdata.Results) == 0 || len(rawdata.Results[0].Series) == 0 {
		err = fmt.Errorf("influxdb return empty value")
		return
	}
	data = rawdata.Results[0].Series[0].Values
	return
}

func (c *Client) LastValue(measurement, tag, field string) (data map[string]Value, err error) {
	query := fmt.Sprintf(`select last(%s) from %s group by "%s"`, field, measurement, tag)
	rawdata, err := c.Exec(query)
	if err != nil {
		return
	}
	if len(rawdata.Results) == 0 {
		err = fmt.Errorf("influxdb return empty value")
		return
	}
	results := rawdata.Results[0]
	data = make(map[string]Value, len(results.Series))
	for _, serial := range results.Series {
		if serial.Values == nil || len(serial.Values) == 0 {
			log.Error("influxdb return empty value")
			continue
		} else {
			values := serial.Values[0]
			if len(values) < 2 {
				log.Error("influxdb return invalid value")
				continue
			}
			timestamp, _ := time.Parse(time.RFC3339Nano, values[0].(string))
			value := values[1]
			data[serial.Tags.Name] = Value{
				Value:     value.(float64),
				Timestamp: timestamp,
			}
			log.Errorf("mirror: %v, progress: %v", serial.Tags.Name, value.(float64))
		}
	}
	return
}

type Data struct {
	Results []struct {
		Series []struct {
			Tags struct {
				Name string `json:"name"`
			} `json:"tags"`
			Values [][]interface{} `json:"values"`
		} `json:"series"`
	} `json:"results"`
}
