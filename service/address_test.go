package service

import (
	"errors"
	"reflect"
	"testing"

	"cmf/paint_proj/model"
)

// mockAddressRepo implements repository.AddressRepository for testing.
// Define only the methods invoked by the service; others can be no-ops as needed.
type mockAddressRepo struct {
	// Capture inputs
	getByUserIDIn          int64
	getAddressListUserIDIn int64
	getAddressListUserName string

	createdIn   *model.Address
	updatedIDIn int64
	updatedMap  map[string]interface{}
	deletedIDIn int64
	setDefUID   int64
	setDefAID   int64

	// Stubbed returns / behavior
	getByUserIDOut   []*model.AddressWithUser // or []model.Address depending on repository; adjust via helper builders below
	getByUserIDErr   error
	getByIDOut       *model.Address
	getByIDErr       error
	getAddrListOut   []*model.AddressWithUser
	getAddrListErr   error
	createErr        error
	updateErr        error
	deleteErr        error
	setDefaultErr    error
}

func (m *mockAddressRepo) GetByUserId(userID int64) ([]*model.AddressWithUser, error) {
	m.getByUserIDIn = userID
	return m.getByUserIDOut, m.getByUserIDErr
}
func (m *mockAddressRepo) GetById(id int64) (*model.Address, error) {
	// optional depending on interface, used in UpdateAdminAddress
	return m.getByIDOut, m.getByIDErr
}
func (m *mockAddressRepo) GetAddressListByUser(userID int64, userName string) ([]*model.AddressWithUser, error) {
	m.getAddressListUserIDIn = userID
	m.getAddressListUserName = userName
	return m.getAddrListOut, m.getAddrListErr
}
func (m *mockAddressRepo) Create(addr *model.Address) error {
	// copy value to avoid external mutation
	if addr \!= nil {
		cp := *addr
		m.createdIn = &cp
	}
	return m.createErr
}
func (m *mockAddressRepo) Update(id int64, updates map[string]interface{}) error {
	m.updatedIDIn = id
	// shallow copy for safety
	m.updatedMap = map[string]interface{}{}
	for k, v := range updates {
		m.updatedMap[k] = v
	}
	return m.updateErr
}
func (m *mockAddressRepo) Delete(id int64) error {
	m.deletedIDIn = id
	return m.deleteErr
}
func (m *mockAddressRepo) SetDefault(userID, addressID int64) error {
	m.setDefUID = userID
	m.setDefAID = addressID
	return m.setDefaultErr
}

/*
Helper builders for model data used in tests.
Adjust field names/types to match actual model structs; we infer from service usage:

- model.Address has: ID, UserId, UserName?, RecipientName, RecipientPhone, Province, City, District, Detail, IsDefault (int), IsDelete (int)
- model.AddressWithUser likely aggregates Address + UserName based on service usage (UserName is read in GetAdminAddressList/AdminGetAddressList).
- model.AddressInfo and model.AdminAddressInfo are DTOs used as service outputs.
- Create/Update request wrappers contain a Data field for Create/Update (except Admin* which flatten fields).
*/
func makeAddrWithUser(id, userID int64, userName, rn, rp, prov, city, dist, detail string, isDefault int) *model.AddressWithUser {
	return &model.AddressWithUser{
		ID:            id,
		UserId:        userID,
		UserName:      userName,
		RecipientName: rn,
		RecipientPhone: rp,
		Province:      prov,
		City:          city,
		District:      dist,
		Detail:        detail,
		IsDefault:     isDefault,
	}
}

func TestGetAddressList_SuccessAndMapping(t *testing.T) {
	repo := &mockAddressRepo{
		getByUserIDOut: []*model.AddressWithUser{
			makeAddrWithUser(1, 10, "alice", "A", "111", "P", "C", "D", "X", 1),
			makeAddrWithUser(2, 10, "alice", "B", "222", "P2", "C2", "D2", "Y", 0),
		},
	}
	svc := NewAddressService(repo)

	got, err := svc.GetAddressList(10)
	if err \!= nil {
		t.Fatalf("GetAddressList returned error: %v", err)
	}
	if repo.getByUserIDIn \!= 10 {
		t.Fatalf("expected repo.GetByUserId called with 10, got %d", repo.getByUserIDIn)
	}
	if len(got) \!= 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	// Verify field mapping and default bool conversion
	if got[0].AddressID \!= 1 || got[0].RecipientName \!= "A" || got[0].RecipientPhone \!= "111" {
		t.Errorf("unexpected first item mapping: %+v", *got[0])
	}
	if got[0].IsDefault == nil || *got[0].IsDefault \!= true {
		t.Errorf("expected first item IsDefault true, got %+v", got[0].IsDefault)
	}
	if got[1].IsDefault == nil || *got[1].IsDefault \!= false {
		t.Errorf("expected second item IsDefault false, got %+v", got[1].IsDefault)
	}
}

func TestGetAddressList_RepoError(t *testing.T) {
	expErr := errors.New("db down")
	repo := &mockAddressRepo{getByUserIDErr: expErr}
	svc := NewAddressService(repo)
	out, err := svc.GetAddressList(5)
	if err == nil || \!errors.Is(err, expErr) {
		t.Fatalf("expected error %v, got %v", expErr, err)
	}
	if out \!= nil {
		t.Fatalf("expected nil out on error, got %#v", out)
	}
}

func TestCreateAddress_DefaultFlagHandling(t *testing.T) {
	trueVal := true
	falseVal := false

	cases := []struct {
		name       string
		isDefault  *bool
		wantStore  int
	}{
		{"nil defaults to 0", nil, 0},
		{"true => 1", &trueVal, 1},
		{"false => 0", &falseVal, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &mockAddressRepo{}
			svc := NewAddressService(repo)
			req := &model.CreateAddressReq{
				Data: model.CreateAddressReqData{
					RecipientName:  "Tom",
					RecipientPhone: "333",
					Province:       "P",
					City:           "C",
					District:       "D",
					Detail:         "addr",
					IsDefault:      tc.isDefault,
				},
			}
			if err := svc.CreateAddress(77, req); err \!= nil {
				t.Fatalf("CreateAddress error: %v", err)
			}
			if repo.createdIn == nil {
				t.Fatalf("expected Create to be called")
			}
			if repo.createdIn.UserId \!= 77 {
				t.Errorf("expected UserId 77, got %d", repo.createdIn.UserId)
			}
			if repo.createdIn.IsDefault \!= tc.wantStore {
				t.Errorf("expected IsDefault %d, got %d", tc.wantStore, repo.createdIn.IsDefault)
			}
		})
	}
}

func TestSetDefaultAddress_Delegates(t *testing.T) {
	repo := &mockAddressRepo{}
	svc := NewAddressService(repo)
	err := svc.SetDefaultAddress(88, 999)
	if err \!= nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.setDefUID \!= 88 || repo.setDefAID \!= 999 {
		t.Fatalf("SetDefault called with wrong args: uid=%d aid=%d", repo.setDefUID, repo.setDefAID)
	}
}

func TestUpdateAddress_FieldFilteringAndDefaultFlag(t *testing.T) {
	trueVal := true
	req := &model.UpdateAddressReq{
		Data: model.UpdateAddressReqData{
			RecipientName:  "NN",
			RecipientPhone: "",
			Province:       "PP",
			City:           "CC",
			District:       "",
			Detail:         "DD",
			IsDefault:      &trueVal, // triggers is_default=1
		},
	}
	repo := &mockAddressRepo{}
	svc := NewAddressService(repo)
	err := svc.UpdateAddress(123, 456, req)
	if err \!= nil {
		t.Fatalf("UpdateAddress error: %v", err)
	}
	if repo.updatedIDIn \!= 456 {
		t.Fatalf("expected update id 456, got %d", repo.updatedIDIn)
	}
	want := map[string]interface{}{
		"recipient_name":  "NN",
		"province":        "PP",
		"city":            "CC",
		"detail":          "DD",
		"is_default":      1,
	}
	if \!reflect.DeepEqual(repo.updatedMap, want) {
		t.Fatalf("unexpected update map.\nwant: %#v\ngot:  %#v", want, repo.updatedMap)
	}
}

func TestUpdateAddress_IsDefaultNilSetsZero(t *testing.T) {
	req := &model.UpdateAddressReq{
		Data: model.UpdateAddressReqData{
			RecipientName: "A",
			IsDefault:     nil, // service sets 0 per code path
		},
	}
	repo := &mockAddressRepo{}
	svc := NewAddressService(repo)
	if err := svc.UpdateAddress(1, 2, req); err \!= nil {
		t.Fatalf("err: %v", err)
	}
	if repo.updatedMap["is_default"] \!= 0 {
		t.Fatalf("expected is_default 0 when nil, got %#v", repo.updatedMap["is_default"])
	}
}

func TestDeleteAddress_Delegates(t *testing.T) {
	repo := &mockAddressRepo{}
	svc := NewAddressService(repo)
	if err := svc.DeleteAddress(100, 200); err \!= nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if repo.deletedIDIn \!= 200 {
		t.Fatalf("expected delete id 200, got %d", repo.deletedIDIn)
	}
}

// Admin (legacy) methods

func TestGetAdminAddressList_MapsFields(t *testing.T) {
	repo := &mockAddressRepo{
		getAddrListOut: []*model.AddressWithUser{
			makeAddrWithUser(1, 7, "root", "R", "123", "P", "C", "D", "det", 1),
		},
	}
	svc := NewAddressService(repo)
	got, err := svc.GetAdminAddressList(7, "root")
	if err \!= nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) \!= 1 {
		t.Fatalf("want 1, got %d", len(got))
	}
	if got[0].AddressID \!= 1 || got[0].UserID \!= 7 || got[0].UserName \!= "root" {
		t.Fatalf("mapping mismatch: %#v", got[0])
	}
	if got[0].IsDefault \!= true {
		t.Fatalf("expected IsDefault true, got %v", got[0].IsDefault)
	}
	if got[0].CreatedAt \!= "" {
		t.Fatalf("expected CreatedAt empty string, got %q", got[0].CreatedAt)
	}
}

func TestCreateAdminAddress_UserIdRequired(t *testing.T) {
	svc := NewAddressService(&mockAddressRepo{})
	err := svc.CreateAdminAddress(0, &model.CreateAddressReq{Data: model.CreateAddressReqData{}})
	if err == nil {
		t.Fatalf("expected error when userId=0")
	}
}

func TestCreateAdminAddress_PassesFieldsAndDefaultLogic(t *testing.T) {
	trueVal := true
	repo := &mockAddressRepo{}
	svc := NewAddressService(repo)
	req := &model.CreateAddressReq{
		Data: model.CreateAddressReqData{
			RecipientName:  "X",
			RecipientPhone: "Y",
			Province:       "P",
			City:           "C",
			District:       "D",
			Detail:         "det",
			IsDefault:      &trueVal,
		},
	}
	if err := svc.CreateAdminAddress(9, req); err \!= nil {
		t.Fatalf("err: %v", err)
	}
	if repo.createdIn == nil || repo.createdIn.UserId \!= 9 || repo.createdIn.IsDefault \!= 1 {
		t.Fatalf("unexpected created: %#v", repo.createdIn)
	}
}

func TestUpdateAdminAddress_NotFoundChecks(t *testing.T) {
	repo := &mockAddressRepo{
		getByIDOut: nil,
		getByIDErr: nil,
	}
	svc := NewAddressService(repo)
	err := svc.UpdateAdminAddress(999, &model.UpdateAddressReq{Data: model.UpdateAddressReqData{RecipientName: "Z"}})
	if err == nil {
		t.Fatalf("expected error when address not found")
	}
}

func TestUpdateAdminAddress_UpdatesProvidedFieldsOnly(t *testing.T) {
	repo := &mockAddressRepo{
		getByIDOut: &model.Address{ID: 3},
	}
	svc := NewAddressService(repo)
	req := &model.UpdateAddressReq{
		Data: model.UpdateAddressReqData{
			RecipientName:  "Z",
			Province:       "P2",
			City:           "C2",
			Detail:         "Det2",
			// IsDefault omitted -> service leaves is_default untouched (no key)
		},
	}
	if err := svc.UpdateAdminAddress(3, req); err \!= nil {
		t.Fatalf("err: %v", err)
	}
	wantSubset := map[string]interface{}{
		"recipient_name": "Z",
		"province":       "P2",
		"city":           "C2",
		"detail":         "Det2",
	}
	// Ensure no "is_default" key added when nil
	if _, ok := repo.updatedMap["is_default"]; ok {
		t.Fatalf("did not expect is_default to be set when nil")
	}
	for k, v := range wantSubset {
		if \!reflect.DeepEqual(repo.updatedMap[k], v) {
			t.Fatalf("key %s mismatch: want %#v got %#v", k, v, repo.updatedMap[k])
		}
	}
	if repo.updatedIDIn \!= 3 {
		t.Fatalf("expected update id 3")
	}
}

// New Admin* methods with pagination and default handling

func TestAdminGetAddressList_PaginationBoundsAndMapping(t *testing.T) {
	// Build 15 entries user 42
	var all []*model.AddressWithUser
	for i := 1; i <= 15; i++ {
		all = append(all, makeAddrWithUser(int64(i), 42, "bob", "R", "P", "P", "C", "D", "det", 0))
	}
	repo := &mockAddressRepo{getAddrListOut: all}
	as := &addressService{addressRepo: repo}

	// default page/pageSize when <=0 should become 1 and 10; here pass -1 to trigger defaults
	got, total, page, pageSize, err := as.AdminGetAddressList(42, "bob", -1, 0)
	if err \!= nil {
		t.Fatalf("err: %v", err)
	}
	if total \!= 15 {
		t.Fatalf("want total 15, got %d", total)
	}
	if page \!= 1 || pageSize \!= 10 {
		t.Fatalf("expected defaults page=1 pageSize=10, got page=%d size=%d", page, pageSize)
	}
	if len(got) \!= 10 {
		t.Fatalf("first page should return 10, got %d", len(got))
	}
	if got[0].AddressID \!= 1 || got[9].AddressID \!= 10 {
		t.Fatalf("unexpected slice range: first=%d last=%d", got[0].AddressID, got[9].AddressID)
	}

	// Second page
	got2, total2, page2, size2, err := as.AdminGetAddressList(42, "bob", 2, 10)
	if err \!= nil {
		t.Fatalf("err: %v", err)
	}
	if total2 \!= 15 || page2 \!= 2 || size2 \!= 10 {
		t.Fatalf("meta mismatch")
	}
	if len(got2) \!= 5 || got2[0].AddressID \!= 11 || got2[4].AddressID \!= 15 {
		t.Fatalf("unexpected second page slice")
	}

	// Out-of-range offset returns empty slice, not error
	got3, total3, _, _, err := as.AdminGetAddressList(42, "bob", 3, 10)
	if err \!= nil {
		t.Fatalf("err: %v", err)
	}
	if total3 \!= 15 {
		t.Fatalf("total mismatch")
	}
	if len(got3) \!= 0 {
		t.Fatalf("expected empty slice when offset >= len, got %d", len(got3))
	}
}

func TestAdminCreateAddress_DefaultFlowCallsSetDefaultThenCreate(t *testing.T) {
	repo := &mockAddressRepo{}
	as := &addressService{addressRepo: repo}

	req := &model.AdminCreateAddressRequest{
		UserID:         50,
		RecipientName:  "M",
		RecipientPhone: "130",
		Province:       "P",
		City:           "C",
		District:       "D",
		Detail:         "det",
		IsDefault:      true,
	}
	if err := as.AdminCreateAddress(req); err \!= nil {
		t.Fatalf("err: %v", err)
	}
	if repo.setDefUID \!= 50 || repo.setDefAID \!= 0 {
		t.Fatalf("expected SetDefault called to clear defaults (aid=0). got uid=%d aid=%d", repo.setDefUID, repo.setDefAID)
	}
	if repo.createdIn == nil || repo.createdIn.IsDefault \!= 1 {
		t.Fatalf("expected created address IsDefault=1, got %#v", repo.createdIn)
	}
}

func TestAdminCreateAddress_PropagatesSetDefaultError(t *testing.T) {
	repo := &mockAddressRepo{setDefaultErr: errors.New("fail clear")}
	as := &addressService{addressRepo: repo}
	req := &model.AdminCreateAddressRequest{UserID: 1, IsDefault: true}
	err := as.AdminCreateAddress(req)
	if err == nil || err.Error() \!= "fail clear" {
		t.Fatalf("expected setDefault error, got %v", err)
	}
}

func TestAdminUpdateAddress_SetsIsDefaultAndClearsExisting(t *testing.T) {
	repo := &mockAddressRepo{}
	as := &addressService{addressRepo: repo}
	req := &model.AdminUpdateAddressRequest{
		ID:            7,
		UserID:        77,
		RecipientName: "N",
		RecipientPhone: "P",
		Province:      "PR",
		City:          "CI",
		District:      "DI",
		Detail:        "DE",
		IsDefault:     true,
	}
	if err := as.AdminUpdateAddress(req); err \!= nil {
		t.Fatalf("err: %v", err)
	}
	// When IsDefault true, SetDefault called with (userID, 0) to clear others
	if repo.setDefUID \!= 77 || repo.setDefAID \!= 0 {
		t.Fatalf("expected clear defaults prior to update, got uid=%d aid=%d", repo.setDefUID, repo.setDefAID)
	}
	if v, ok := repo.updatedMap["is_default"]; \!ok || v.(int) \!= 1 {
		t.Fatalf("expected is_default=1 in update map, got %#v", repo.updatedMap["is_default"])
	}
	if repo.updatedIDIn \!= 7 {
		t.Fatalf("expected update id 7")
	}
}

func TestAdminUpdateAddress_IsDefaultFalseSetsZero(t *testing.T) {
	repo := &mockAddressRepo{}
	as := &addressService{addressRepo: repo}
	req := &model.AdminUpdateAddressRequest{
		ID:        8,
		UserID:    88,
		IsDefault: false,
	}
	if err := as.AdminUpdateAddress(req); err \!= nil {
		t.Fatalf("err: %v", err)
	}
	if repo.updatedMap["is_default"] \!= 0 {
		t.Fatalf("expected is_default=0 when false")
	}
}

func TestAdminDeleteAddress_Delegates(t *testing.T) {
	repo := &mockAddressRepo{}
	as := &addressService{addressRepo: repo}
	if err := as.AdminDeleteAddress(1234); err \!= nil {
		t.Fatalf("err: %v", err)
	}
	if repo.deletedIDIn \!= 1234 {
		t.Fatalf("expected delete id 1234, got %d", repo.deletedIDIn)
	}
}