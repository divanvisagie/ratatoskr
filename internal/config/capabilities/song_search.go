package capabilities

import "github.com/divanvisagie/ratatoskr/pkg/types"

type SongSearchCapability struct {
	out chan types.ResponseMessage	
}

func NewSongSearchCapability() *SongSearchCapability {
	return &SongSearchCapability{}
}
