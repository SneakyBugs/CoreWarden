package provider

import (
	"context"
	"testing"
	"time"

	"git.houseofkummer.com/lior/home-dns/client"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

type MockClientAction struct {
	action     any
	stateAfter []client.Record
}

type MockClient struct {
	actions            []MockClientAction
	currentActionIndex int
	t                  *testing.T
}

func (c *MockClient) CreateRecord(params client.CreateRecordParams) (client.Record, error) {
	if len(c.actions) <= c.currentActionIndex {
		c.t.Fatalf("Client called CreateRecord when no more method calls were expected\n")
	}
	clientAction, ok := c.actions[c.currentActionIndex].action.(CreateRecordAction)
	if !ok {
		c.t.Fatalf("Client called unexpected method CreateRecord during action %d\n", c.currentActionIndex)
	}
	if clientAction.Zone != params.Zone {
		c.t.Fatalf("Expected params.Zone to be '%s', got '%s'\n", clientAction.Zone, params.Zone)
	}
	if clientAction.RR != params.RR {
		c.t.Fatalf("Expected params.RR to be '%s', got '%s'\n", clientAction.RR, params.RR)
	}
	if clientAction.Comment != params.Comment {
		c.t.Fatalf("Expected params.Comment to be '%s', got '%s'\n", clientAction.Comment, params.Comment)
	}
	c.currentActionIndex += 1
	return clientAction.ResponseRecord, clientAction.ResponseErr
}

type CreateRecordAction struct {
	Zone           string
	RR             string
	Comment        string
	ResponseRecord client.Record
	ResponseErr    error
}

func (c *MockClient) ReadRecord(id int) (client.Record, error) {
	if len(c.actions) <= c.currentActionIndex {
		c.t.Fatalf("Client called ReadRecord when no more method calls were expected\n")
	}
	clientAction, ok := c.actions[c.currentActionIndex].action.(ReadRecordAction)
	if !ok {
		c.t.Fatalf("Client called unexpected method ReadRecord during action %d\n", c.currentActionIndex)
	}
	if clientAction.ID != id {
		c.t.Fatalf("Expected id to be '%d', got '%d'\n", clientAction.ID, id)
	}
	c.currentActionIndex += 1
	return clientAction.ResponseRecord, clientAction.ResponseErr
}

type ReadRecordAction struct {
	ID             int
	ResponseRecord client.Record
	ResponseErr    error
}

func (c *MockClient) UpdateRecord(params client.UpdateRecordParams) (client.Record, error) {
	if len(c.actions) <= c.currentActionIndex {
		c.t.Fatalf("Client called UpdateRecord when no more method calls were expected\n")
	}
	clientAction, ok := c.actions[c.currentActionIndex].action.(UpdateRecordAction)
	if !ok {
		c.t.Fatalf("Client called unexpected method UpdateRecord during action %d\n", c.currentActionIndex)
	}
	if clientAction.ID != params.ID {
		c.t.Fatalf("Expected params.ID to be '%d', got '%d\n", clientAction.ID, params.ID)
	}
	if clientAction.Zone != params.Zone {
		c.t.Fatalf("Expected params.Zone to be '%s', got '%s'\n", clientAction.Zone, params.Zone)
	}
	if clientAction.RR != params.RR {
		c.t.Fatalf("Expected params.RR to be '%s', got '%s'\n", clientAction.RR, params.RR)
	}
	if clientAction.Comment != params.Comment {
		c.t.Fatalf("Expected params.Comment to be '%s', got '%s'\n", clientAction.Comment, params.Comment)
	}
	c.currentActionIndex += 1
	return clientAction.ResponseRecord, clientAction.ResponseErr
}

type UpdateRecordAction struct {
	ID             int
	Zone           string
	RR             string
	Comment        string
	ResponseRecord client.Record
	ResponseErr    error
}

func (c *MockClient) DeleteRecord(id int) (client.Record, error) {
	if len(c.actions) <= c.currentActionIndex {
		c.t.Fatalf("Client called DeleteRecord when no more method calls were expected\n")
	}
	clientAction, ok := c.actions[c.currentActionIndex].action.(DeleteRecordAction)
	if !ok {
		c.t.Fatalf("Client called unexpected method DeleteRecord during action %d\n", c.currentActionIndex)
	}
	if clientAction.ID != id {
		c.t.Fatalf("Expected id to be '%d', got '%d'\n", clientAction.ID, id)
	}
	c.currentActionIndex += 1
	return clientAction.ResponseRecord, clientAction.ResponseErr
}

type DeleteRecordAction struct {
	ID             int
	ResponseRecord client.Record
	ResponseErr    error
}

func (c *MockClient) ListRecords(zone string) ([]client.Record, error) {
	if len(c.actions) <= c.currentActionIndex {
		c.t.Fatalf("Client called ListRecords when no more method calls were expected\n")
	}
	clientAction, ok := c.actions[c.currentActionIndex].action.(ListRecordAction)
	if !ok {
		c.t.Fatalf("Client called unexpected method ListRecords during action %d\n", c.currentActionIndex)
	}
	if clientAction.Zone != zone {
		c.t.Fatalf("Expected zone to be '%s', got '%s'\n", clientAction.Zone, zone)
	}
	c.currentActionIndex += 1
	return clientAction.ResponseRecords, clientAction.ResponseErr
}

type ListRecordAction struct {
	Zone            string
	ResponseRecords []client.Record
	ResponseErr     error
}

func newTestProvider(t *testing.T, actions []MockClientAction) Provider {
	mockClient := MockClient{
		actions:            actions,
		currentActionIndex: 0,
		t:                  t,
	}
	return Provider{
		client: &mockClient,
		zones:  []string{"example.com"},
	}
}

func createTestRecord(t *testing.T, id int, zone string, content string, comment string) client.Record {
	now := time.Now()
	rr, err := dns.NewRR(content)
	assert.Nil(t, err)
	return client.Record{
		ID:        id,
		Zone:      zone,
		RR:        rr,
		Comment:   comment,
		CreatedAt: now,
		UpdatedOn: now,
	}
}

func TestRecordsBasic(t *testing.T) {
	state := []client.Record{
		createTestRecord(t, 1, "example.com.", ". IN A 10.0.0.1", ""),
		createTestRecord(t, 2, "example.com.", "@ IN A 10.0.0.2", ""),
		createTestRecord(t, 2, "example.com.", "foo IN A 10.0.0.3", ""),
	}
	p := newTestProvider(
		t,
		[]MockClientAction{
			{
				action: ListRecordAction{
					Zone:            "example.com",
					ResponseRecords: state,
					ResponseErr:     nil,
				},
				stateAfter: state,
			},
		},
	)
	endpoints, err := p.Records(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 2, len(endpoints), "endpoints length should be 2")
	assert.Equal(t, 2, len(endpoints[0].Targets), "endpoints[0].Targets length should be 2")
	assert.Equal(t, "10.0.0.1", endpoints[0].Targets[0])
	assert.Equal(t, "10.0.0.2", endpoints[0].Targets[1])
	assert.Equal(t, "example.com", endpoints[0].DNSName)
	assert.Equal(t, 1, len(endpoints[1].Targets), "endpoints[1].Targets length should be 1")
	assert.Equal(t, "10.0.0.3", endpoints[1].Targets[0])
	assert.Equal(t, "foo.example.com", endpoints[1].DNSName)
}

func TestRecordsIgnoresUnsupportedTypes(t *testing.T) {
	state := []client.Record{
		createTestRecord(t, 1, "example.com.", ". IN A 10.0.0.1", ""),
		createTestRecord(t, 2, "example.com.", ". IN MX 10 mail.example.com.", ""),
	}
	p := newTestProvider(
		t,
		[]MockClientAction{
			{
				action: ListRecordAction{
					Zone:            "example.com",
					ResponseRecords: state,
					ResponseErr:     nil,
				},
				stateAfter: state,
			},
		},
	)
	endpoints, err := p.Records(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(endpoints), "endpoints length should be 2")
	assert.Equal(t, 1, len(endpoints[0].Targets), "endpoints[0].Targets length should be 2")
	assert.Equal(t, "10.0.0.1", endpoints[0].Targets[0])
	assert.Equal(t, "example.com", endpoints[0].DNSName)
}
