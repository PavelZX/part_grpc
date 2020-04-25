package part

import (
	"context"

	part "github.com/abronan/part-grpc/api/part/v1"
	"github.com/go-pg/pg"
	"github.com/gogo/protobuf/types"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Service is the service dealing with storing
// and retrieving part items from the database.
type Service struct {
	DB *pg.DB
}

// CreatePart creates a part given a description
func (s Service) CreatePart(ctx context.Context, req *part.CreatePartRequest) (*part.CreatePartResponse, error) {
	req.Item.Id = uuid.NewV4().String()
	err := s.DB.Insert(req.Item)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not insert item into the database: %s", err)
	}
	return &part.CreatePartResponse{Id: req.Item.Id}, nil
}

// CreateParts create part items from a list of part descriptions
func (s Service) CreateParts(ctx context.Context, req *part.CreatePartsRequest) (*part.CreatePartsResponse, error) {
	var ids []string
	for _, item := range req.Items {
		item.Id = uuid.NewV4().String()
		ids = append(ids, item.Id)
	}
	err := s.DB.Insert(&req.Items)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not insert items into the database: %s", err)
	}
	return &part.CreatePartsResponse{Ids: ids}, nil
}

// GetPart retrieves a part item from its ID
func (s Service) GetPart(ctx context.Context, req *part.GetPartRequest) (*part.GetPartResponse, error) {
	var item part.Part
	err := s.DB.Model(&item).Where("id = ?", req.Id).First()
	if err != nil {
		return nil, grpc.Errorf(codes.NotFound, "Could not retrieve item from the database: %s", err)
	}
	return &part.GetPartResponse{Item: &item}, nil
}

// ListPart retrieves a part item from its ID
func (s Service) ListPart(ctx context.Context, req *part.ListPartRequest) (*part.ListPartResponse, error) {
	var items []*part.Part
	query := s.DB.Model(&items).Order("created_at ASC")
	if req.Limit > 0 {
		query.Limit(int(req.Limit))
	}
	if req.NotCompleted {
		query.Where("completed = false")
	}
	err := query.Select()
	if err != nil {
		return nil, grpc.Errorf(codes.NotFound, "Could not list items from the database: %s", err)
	}
	return &part.ListPartResponse{Items: items}, nil
}

// DeletePart deletes a part given an ID
func (s Service) DeletePart(ctx context.Context, req *part.DeletePartRequest) (*part.DeletePartResponse, error) {
	err := s.DB.Delete(&part.Part{Id: req.Id})
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not delete item from the database: %s", err)
	}
	return &part.DeletePartResponse{}, nil
}

// UpdatePart updates a part item
func (s Service) UpdatePart(ctx context.Context, req *part.UpdatePartRequest) (*part.UpdatePartResponse, error) {
	req.Item.UpdatedAt = types.TimestampNow()
	res, err := s.DB.Model(req.Item).Column("title", "description", "completed", "updated_at").Update()
	if res.RowsAffected() == 0 {
		return nil, grpc.Errorf(codes.NotFound, "Could not update item: not found")
	}
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not update item from the database: %s", err)
	}
	return &part.UpdatePartResponse{}, nil
}

// UpdateParts updates part items given their respective title and description.
func (s Service) UpdateParts(ctx context.Context, req *part.UpdatePartsRequest) (*part.UpdatePartsResponse, error) {
	time := types.TimestampNow()
	for _, item := range req.Items {
		item.UpdatedAt = time
	}
	res, err := s.DB.Model(&req.Items).Column("title", "description", "completed", "updated_at").Update()
	if res.RowsAffected() == 0 {
		return nil, grpc.Errorf(codes.NotFound, "Could not update items: not found")
	}
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, "Could not update items from the database: %s", err)
	}
	return &part.UpdatePartsResponse{}, nil
}
