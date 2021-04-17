package truenas

import (
	"context"
	"fmt"
	"net/url"
)

// PoolService handles communication with the dataset related
// methods of the TrueNAS API.
type DatasetService service

type CreateDatasetInput struct {
	ATime             string `json:"atime,omitempty"`
	ACLMode           string `json:"aclmode,omitempty"`
	Name              string `json:"name"`
	Comments          string `json:"comments,omitempty"`
	Compression       string `json:"compression,omitempty"`
	CaseSensitivity   string `json:"casesensitivity,omitempty"`
	Copies            int    `json:"copies,omitempty"`
	Deduplication     string `json:"deduplication,omitempty"`
	InheritEncryption *bool  `json:"inherit_encryption,omitempty"`
	Quota             *int64 `json:"quota,omitempty"`
	RefQuota          *int64 `json:"refquota,omitempty"`
	RefReservation    *int64 `json:"refreservation,omitempty"`
	Reservation       *int64 `json:"reservation,omitempty"`
	ShareType         string `json:"share_type,omitempty"`
	Sync              string `json:"sync,omitempty"`
}

// CompositeValue composite value type that most TrueNAS seem to be using
type CompositeValue struct {
	Value    *string `json:"value"`
	RawValue string  `json:"rawvalue"`
	//Parsed   string  `json:"parsed"` // looks like TrueNAS mixes types for this property: bool/string/number?
	Source string `json:"source"`
}

type DatasetResponse struct {
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	Comments              *CompositeValue `json:"comments"`
	Pool                  string          `json:"pool"`
	Type                  string          `json:"type"`
	Mountpoint            string          `json:"mountpoint"`
	Encrypted             bool            `json:"encrypted"`
	KeyLoaded             bool            `json:"key_loaded"`
	ManagedBy             *CompositeValue `json:"managedby"`
	Deduplication         *CompositeValue `json:"deduplication"`
	ACLMode               *CompositeValue `json:"aclmode"`
	ACLType               *CompositeValue `json:"acltype"`
	XATTR                 *CompositeValue `json:"xattr"`
	ATime                 *CompositeValue `json:"atime"`
	CaseSensitivity       *CompositeValue `json:"casesensitivity"`
	Exec                  *CompositeValue `json:"exec"`
	Sync                  *CompositeValue `json:"sync"`
	Compression           *CompositeValue `json:"compression"`
	CompressRatio         *CompositeValue `json:"compressratio"`
	Origin                *CompositeValue `json:"origin"`
	Quota                 *CompositeValue `json:"quota"`
	RefQuota              *CompositeValue `json:"refquota"`
	Reservation           *CompositeValue `json:"reservation"`
	RefReservation        *CompositeValue `json:"refreservation"`
	Copies                *CompositeValue `json:"copies"`
	SnapDir               *CompositeValue `json:"snapdir"`
	Readonly              *CompositeValue `json:"readonly"`
	Recordsize            *CompositeValue `json:"recordsize"`
	KeyFormat             *CompositeValue `json:"key_format"`
	EncryptionAlgorithm   *CompositeValue `json:"encryption_algorithm"`
	Used                  *CompositeValue `json:"used"`
	Available             *CompositeValue `json:"available"`
	SpecialSmallBlockSize *CompositeValue `json:"special_small_block_size"`
	PBKDF2Iters           *CompositeValue `json:"pbkdf2iters"`
	Locked                bool            `json:"locked"`
}

func (s *DatasetService) Create(ctx context.Context, dataset *CreateDatasetInput) (*DatasetResponse, error) {
	path := "pool/dataset"

	body := struct {
		CreateDatasetInput
		Type string `json:"type"`
	}{
		CreateDatasetInput: *dataset,
		Type:               "FILESYSTEM",
	}

	req, err := s.client.NewRequest("POST", path, body)

	if err != nil {
		return nil, err
	}

	d := &DatasetResponse{}

	_, err = s.client.Do(ctx, req, d)

	if err != nil {
		return nil, err
	}

	return d, nil
}

func (s *DatasetService) Get(ctx context.Context, id string) (*DatasetResponse, error) {
	path := fmt.Sprintf("pool/dataset/id/%s", url.QueryEscape(id))
	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	dataset := &DatasetResponse{}

	_, err = s.client.Do(ctx, req, dataset)
	if err != nil {
		return nil, err
	}

	return dataset, nil
}

func (s *DatasetService) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("pool/dataset/id/%s", url.QueryEscape(id))

	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}
