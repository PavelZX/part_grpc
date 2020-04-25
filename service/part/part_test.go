package part

import (
	"context"
	"testing"
	"time"

	api "github.com/abronan/part-grpc/api/part/v1"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PartSuite struct {
	suite.Suite
	Part *Service
}

func TestPartTestSuite(t *testing.T) {
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Database: "part",
		Addr:     "localhost:5432",
		RetryStatementTimeout: true,
		MaxRetries:            4,
		MinRetryBackoff:       250 * time.Millisecond,
	})
	suite.Run(t, &PartSuite{
		Part: &Service{DB: db},
	})
}

func (s *PartSuite) SetupTest() {
	s.Part.DB.DropTable(&api.Part{}, &orm.DropTableOptions{IfExists: true})
	s.Part.DB.CreateTable(&api.Part{}, nil)
}

func (s *PartSuite) TearDownTest() {
	s.Part.DB.DropTable(&api.Part{}, &orm.DropTableOptions{IfExists: true})
}

func (s *PartSuite) TestCreatePart() {
	rcreate, err := s.Part.CreatePart(
		context.Background(),
		&api.CreatePartRequest{
			Item: &api.Part{
				Title:       "item_1",
				Description: "item desc 1",
			},
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rcreate)
	assert.NotEqual(s.T(), rcreate.Id, "")
}

func (s *PartSuite) TestCreateParts() {
	rcreate, err := s.Part.CreateParts(
		context.Background(),
		&api.CreatePartsRequest{
			Items: []*api.Part{
				&api.Part{
					Title:       "item_1",
					Description: "item desc 1",
				},
				&api.Part{
					Title:       "item_2",
					Description: "item desc 2",
				},
			},
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rcreate)
	for _, id := range rcreate.Ids {
		assert.NotEqual(s.T(), id, "")
	}
}

func (s *PartSuite) TestGetPart() {
	item := &api.Part{
		Title:       "item_1",
		Description: "item desc 1",
	}

	rcreate, err := s.Part.CreatePart(
		context.Background(),
		&api.CreatePartRequest{
			Item: item,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rcreate)
	assert.NotEqual(s.T(), rcreate.Id, "")

	id := rcreate.Id

	rget, err := s.Part.GetPart(
		context.Background(),
		&api.GetPartRequest{
			Id: id,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rget)
	assert.NotNil(s.T(), rget.Item)
	assert.Equal(s.T(), rget.Item, item)
}

func (s *PartSuite) TestDeletePart() {
	item := &api.Part{
		Title:       "item_1",
		Description: "item desc 1",
	}

	rcreate, err := s.Part.CreatePart(
		context.Background(),
		&api.CreatePartRequest{
			Item: item,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rcreate)
	assert.NotEqual(s.T(), rcreate.Id, "")

	id := rcreate.Id

	rdel, err := s.Part.DeletePart(
		context.Background(),
		&api.DeletePartRequest{
			Id: id,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rdel)

	// Getting the part item should fail this time
	rget, err := s.Part.GetPart(
		context.Background(),
		&api.GetPartRequest{
			Id: id,
		},
	)
	assert.Nil(s.T(), rget)
	assert.NotNil(s.T(), err)
	assert.Contains(s.T(), err.Error(), "Could not retrieve item from the database: pg: no rows in result set")
}

func (s *PartSuite) TestUpdatePart() {
	item := &api.Part{
		Title:       "item_1",
		Description: "item desc 1",
	}

	rcreate, err := s.Part.CreatePart(
		context.Background(),
		&api.CreatePartRequest{
			Item: item,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rcreate)
	assert.NotEqual(s.T(), rcreate.Id, "")

	id := rcreate.Id

	newItem := &api.Part{
		Id:          id,
		Title:       "item 1 update",
		Description: "updated desc",
		Completed:   true,
	}

	rupdate, err := s.Part.UpdatePart(
		context.Background(),
		&api.UpdatePartRequest{
			Item: newItem,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rupdate)

	// Getting the part item should return the updated version
	rget, err := s.Part.GetPart(
		context.Background(),
		&api.GetPartRequest{
			Id: id,
		},
	)
	assert.NotNil(s.T(), rget)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), rget.Item.Id, newItem.Id)
	assert.Equal(s.T(), rget.Item.Title, newItem.Title)
	assert.Equal(s.T(), rget.Item.Description, newItem.Description)
	assert.Equal(s.T(), rget.Item.Completed, newItem.Completed)
}

func (s *PartSuite) TestUpdateParts() {
	items := []*api.Part{
		{
			Title:       "item_1",
			Description: "item desc 1",
		},
		{
			Title:       "item_2",
			Description: "item desc 2",
		},
	}

	// Create the part items
	resp, err := s.Part.CreateParts(
		context.Background(),
		&api.CreatePartsRequest{
			Items: items,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), resp)

	// List the items and update their fields
	rlist, err := s.Part.ListPart(
		context.Background(),
		&api.ListPartRequest{},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rlist)
	assert.NotNil(s.T(), rlist.Items)

	for _, item := range rlist.Items {
		item.Description = "updated desc"
		item.Completed = true
	}

	rupdate, err := s.Part.UpdateParts(
		context.Background(),
		&api.UpdatePartsRequest{
			Items: rlist.Items,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rupdate)

	// List again and see if the entries have had their fields changed
	rlist, err = s.Part.ListPart(
		context.Background(),
		&api.ListPartRequest{},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rlist)
	assert.NotNil(s.T(), rlist.Items)

	for _, item := range rlist.Items {
		assert.Equal(s.T(), item.Description, "updated desc")
		assert.True(s.T(), item.Completed)
	}
}

func (s *PartSuite) TestListPart() {
	items := []*api.Part{
		{
			Title:       "item_1",
			Description: "item desc 1",
			Completed:   true,
		},
		{
			Title:       "item_2",
			Description: "item desc 2",
		},
		{
			Title:       "item_3",
			Description: "item desc 3",
		},
		{
			Title:       "item_4",
			Description: "item desc 4",
			Completed:   true,
		},
	}

	// List with empty database
	rlist, err := s.Part.ListPart(
		context.Background(),
		&api.ListPartRequest{},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rlist)
	assert.Nil(s.T(), rlist.Items)
	assert.Equal(s.T(), len(rlist.Items), 0)

	// Create the part items
	rcreate, err := s.Part.CreateParts(
		context.Background(),
		&api.CreatePartsRequest{
			Items: items,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rcreate)

	// List the items
	rlist, err = s.Part.ListPart(
		context.Background(),
		&api.ListPartRequest{},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rlist)
	assert.NotNil(s.T(), rlist.Items)
	assert.Equal(s.T(), len(rlist.Items), 4)

	// Limit the result of List
	rlist, err = s.Part.ListPart(
		context.Background(),
		&api.ListPartRequest{
			Limit: 2,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rlist)
	assert.NotNil(s.T(), rlist.Items)
	assert.Equal(s.T(), len(rlist.Items), 2)

	// Only list non completed items
	rlist, err = s.Part.ListPart(
		context.Background(),
		&api.ListPartRequest{
			NotCompleted: true,
		},
	)
	assert.Nil(s.T(), err)
	assert.NotNil(s.T(), rlist)
	assert.NotNil(s.T(), rlist.Items)
	assert.Equal(s.T(), len(rlist.Items), 2)
}
