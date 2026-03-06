package grpc

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/hardvlad/ypshort/internal/auth"
	yphandler "github.com/hardvlad/ypshort/internal/handler"
	"github.com/hardvlad/ypshort/internal/service"
	pb "github.com/hardvlad/ypshort/proto"
)

type Server struct {
	pb.UnimplementedShortenerServiceServer
	service *service.ShortenerService
}

func New(svc *service.ShortenerService) *Server {
	return &Server{service: svc}
}

func AuthInterceptor(sugarLogger *zap.SugaredLogger, secretKey string, db *sql.DB) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, service.ErrUnauthorized
		}

		authHeaders := md["authorization"]
		if len(authHeaders) == 0 || authHeaders[0] == "" {
			return nil, service.ErrUnauthorized
		}

		userID, err := auth.GetUserID(authHeaders[0], secretKey)
		if err != nil {
			sugarLogger.Errorw(err.Error(), "event", "парсинг токена из хедера", "header", authHeaders[0])
			return nil, service.ErrUnauthorized
		}

		if userID == 0 {
			userID, _, err = auth.CreateNewUser(db, secretKey)
			if err != nil {
				sugarLogger.Errorw(err.Error(), "event", "создание нового пользователя")
				return nil, service.ErrBadRequest
			}
		}

		ctx = context.WithValue(ctx, yphandler.UserIDKey, userID)
		ctx = metadata.NewOutgoingContext(ctx, md)

		return handler(ctx, req)
	}
}

func (s *Server) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {

	userID, ok := ctx.Value(yphandler.UserIDKey).(int)
	if !ok {
		userID = 0
	}

	success, fullURL, _, err := s.service.Shorten(req.GetUrl(), userID)

	var response pb.URLShortenResponse

	if !success || err != nil {
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	response.SetResult(fullURL)
	return &response, nil
}

func (s *Server) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	var response pb.URLExpandResponse
	urlRedirect, isDeleted, ok := s.service.Expand(req.GetId())
	if isDeleted {
		return nil, status.Errorf(codes.NotFound, "short link was deleted")
	}

	if ok {
		response.SetResult(urlRedirect)
		return &response, nil
	}
	return nil, status.Errorf(codes.NotFound, "short link not found")
}

func (s *Server) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, ok := ctx.Value(yphandler.UserIDKey).(int)
	if !ok {
		userID = 0
	}

	userURLs, err := s.service.ListUserURLs(ctx, userID)
	if err != nil {
		switch err.(type) {
		case *service.APIError:
			return nil, status.Errorf(status.Code(err), "%v", err)
		default:
			return nil, status.Error(codes.Internal, "internal server error")
		}
	}

	var pbUrls []*pb.URLData

	var response pb.UserURLsResponse
	for shortCode, originalURL := range userURLs {
		var d = pb.URLData{}
		d.SetOriginalUrl(originalURL)
		d.SetShortUrl(shortCode)
		pbUrls = append(pbUrls, &d)
	}

	response.SetUrl(pbUrls)

	return &response, nil
}
