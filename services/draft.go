package services

import (
	"database/sql"
	"net/http"
)

type DraftStorageImpl struct {
	DB *sql.DB
}

func (s *DraftStorageImpl) ListDrafts(w http.ResponseWriter, r *http.Request) {
	/*drafts := &pb.ListDraftsResponse{}
	if err := database.List(s.DB, "draft", &drafts, "order by data->'$.created' desc"); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	//draft := &pb.Draft{
	//	Id:       "abcd",
	//	Snipped:  "Hello",
	//	Envelope: nil,
	//}
	//drafts.Draft = append(drafts.Draft, draft)
	//draft = &pb.Draft{
	//	Id:       "efgh",
	//	Snipped:  "World",
	//	Envelope: nil,
	//}
	//drafts.Draft = append(drafts.Draft, draft)

	//response, err := proto.Marshal(drafts)
	//if err != nil {
	//	log.Fatalf("Unable to marshal response : %v", err)
	//}
	//w.Write(response)

	response, err := json.Marshal(drafts)
	if err != nil {
		log.Fatalf("Unable to marshal response : %v", err)
	}
	w.Write(response)*/

	w.Write([]byte("Hello World!"))
}
