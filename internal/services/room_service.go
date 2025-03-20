package services

import (
	"music-share-api/internal/repositories"
)

type RoomService interface {
	CreateRoom(input repositories.RoomCreateInput) (int, error)
}

type roomService struct {
	roomRepository repositories.RoomRepository
}

func NewRoomService(roomRepository repositories.RoomRepository) RoomService {
	return &roomService{
		roomRepository: roomRepository,
	}
}

func (s *roomService) CreateRoom(input repositories.RoomCreateInput) (int, error) {
	return s.roomRepository.CreateRoom(input)
}
