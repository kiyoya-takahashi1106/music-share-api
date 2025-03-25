package services

import (
	"fmt"
	"music-share-api/internal/repositories"
)

type RoomService interface {
	CreateRoom(input repositories.RoomCreateInput) (int, error)
	JoinRoom(userID int, userName string, roomID int, roomPassword *string) (error)
	LeaveRoom(userID int, roomID int) (*repositories.RoomAllInfo, error)
	DeleteRoom(roomID int) error
	GetRoom(roomID int) (*repositories.RoomAllInfo, error)
}

type roomService struct {
	roomRepository repositories.RoomRepository
}

func NewRoomService(roomRepository repositories.RoomRepository) RoomService {
	return &roomService{roomRepository: roomRepository}
}

func (s *roomService) CreateRoom(input repositories.RoomCreateInput) (int, error) {
	roomID, err := s.roomRepository.CreateRoom(input)
	if err != nil {
		return 0, fmt.Errorf("failed to create room: %w", err)
	}
	return roomID, nil
}

func (s *roomService) JoinRoom(userID int, userName string, roomID int, roomPassword *string) (error) {
	// まず参加処理を実行
	if err := s.roomRepository.JoinRoom(userID, userName, roomID, roomPassword); err != nil {
		return fmt.Errorf("failed to join room: %w", err)
	}
	return nil
}

func (s *roomService) LeaveRoom(userID int, roomID int) (*repositories.RoomAllInfo, error) {
	room, err := s.roomRepository.LeaveRoom(userID, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to leave room: %w", err)
	}
	return room, nil
}

func (s *roomService) DeleteRoom(roomID int) error {
	if err := s.roomRepository.DeleteRoom(roomID); err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}
	return nil
}

func (s *roomService) GetRoom(roomID int) (*repositories.RoomAllInfo, error) {
	room, err := s.roomRepository.GetRoomByID(roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room: %w", err)
	}
	return room, nil
}
