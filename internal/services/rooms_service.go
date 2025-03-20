package services

import (
    "music-share-api/internal/repositories"
)

type RoomsService interface {
    GetPublicRooms() ([]repositories.Room, error)
}

type roomsService struct {
    roomsRepository repositories.RoomsRepository
}

func NewRoomsService(roomsRepository repositories.RoomsRepository) RoomsService {
    return &roomsService{
        roomsRepository: roomsRepository,
    }
}

func (s *roomsService) GetPublicRooms() ([]repositories.Room, error) {
    return s.roomsRepository.GetPublicRooms()
}