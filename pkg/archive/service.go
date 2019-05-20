package archive

type ArchiveService interface {
	RecognizeParticipant(name string) uint64
	RecognizeLiveMatchEvent(team1, team2 string) uint64
}
