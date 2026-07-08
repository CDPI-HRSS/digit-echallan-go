package service

import (
	"encoding/json"
	"go.uber.org/zap"
	"fmt"
	"log"

	config "github.com/CDPI-HRSS/echallan-services/configs"
	"github.com/CDPI-HRSS/echallan-services/internal/domain"
	"github.com/CDPI-HRSS/echallan-services/internal/repository/http"
	"github.com/CDPI-HRSS/echallan-services/internal/transport/kafka"
)

type NotificationService struct {
	cfg      *config.Config
	producer *kafka.Producer
	mdmsRepo *http.MdmsRepository
}

func NewNotificationService(cfg *config.Config, producer *kafka.Producer, mdmsRepo *http.MdmsRepository) *NotificationService {
	return &NotificationService{
		cfg:      cfg,
		producer: producer,
		mdmsRepo: mdmsRepo,
	}
}

func (s *NotificationService) SendNotifications(requestInfo *domain.RequestInfo, challan *domain.Challan, action string) {
	if challan.Citizen == nil || challan.Citizen.MobileNumber == "" {
		log.Printf("Cannot send notification: Missing citizen details for challan %s", challan.ChallanNo)
		return
	}

	// Fetch ChannelList from MDMS
	mdmsData, err := s.mdmsRepo.FetchMasterData(requestInfo, challan.TenantId)
	
	sendSMS := true
	sendEmail := false
	sendEvent := true // default to true if MDMS fails

	if err == nil && mdmsData != nil {
		if commonMasters, ok := mdmsData["egov-common-masters"].(map[string]interface{}); ok {
			if channels, ok := commonMasters["ChannelList"].([]interface{}); ok {
				sendSMS, sendEmail, sendEvent = false, false, false
				for _, chRaw := range channels {
					if ch, ok := chRaw.(map[string]interface{}); ok {
						code, _ := ch["code"].(string)
						if code == "SMS" {
							sendSMS = true
						} else if code == "EMAIL" {
							sendEmail = true
						} else if code == "EVENT" {
							sendEvent = true
						}
					}
				}
			}
		}
	} else {
		log.Printf("Warning: Failed to fetch ChannelList from MDMS, using defaults. Error: %v", err)
	}

	msg := fmt.Sprintf("Dear Citizen, your Challan %s has been %s.", challan.ChallanNo, action)

	// 1. SMS
	if sendSMS {
		smsReq := map[string]interface{}{
			"mobileNumber": challan.Citizen.MobileNumber,
			"message":      msg,
		}
		_ = s.producer.Push(s.cfg.SMSTopic, smsReq)
	}

	// 2. Email
	if sendEmail && challan.Citizen.EmailId != "" {
		emailReq := map[string]interface{}{
			"email":   challan.Citizen.EmailId,
			"subject": fmt.Sprintf("eChallan Update: %s", challan.ChallanNo),
			"body":    msg, // Basic text email implementation
			"isHTML":  false,
		}
		_ = s.producer.Push(s.cfg.EmailTopic, emailReq)
	}

	// 3. Event (In-App)
	if sendEvent {
		payLink := fmt.Sprintf("https://digit.org/citizen/payment/my-bills/%s/%s", challan.BusinessService, challan.ChallanNo)
		receiptLink := fmt.Sprintf("https://digit.org/citizen/receipts/%s/%s", challan.BusinessService, challan.ChallanNo)

		actionArr := []map[string]interface{}{}
		if action == "CREATED" {
			actionArr = append(actionArr, map[string]interface{}{
				"tenantId": challan.TenantId,
				"actionUrls": []map[string]interface{}{
					{"actionUrl": payLink, "code": "PAY_NOW"},
				},
			})
		} else if action == "PAID" {
			actionArr = append(actionArr, map[string]interface{}{
				"tenantId": challan.TenantId,
				"actionUrls": []map[string]interface{}{
					{"actionUrl": receiptLink, "code": "DOWNLOAD_RECEIPT"},
				},
			})
		}

		eventReq := map[string]interface{}{
			"events": []map[string]interface{}{
				{
					"tenantId":    challan.TenantId,
					"eventType":   "SYSTEMGENERATED",
					"description": msg,
					"name":        "Challan Update",
					"source":      "WEBAPP",
					"receipient": map[string]interface{}{
						"toUsers": []string{challan.Citizen.Uuid},
					},
					"actions": map[string]interface{}{
						"tenantId": challan.TenantId,
						"actionUrls": func() interface{} {
							if len(actionArr) > 0 {
								return actionArr[0]["actionUrls"]
							}
							return []interface{}{}
						}(),
					},
				},
			},
		}
		_ = s.producer.Push(s.cfg.UserEventTopic, eventReq)
	}

	log.Printf("Successfully pushed notifications for challan %s", challan.ChallanNo)
}

func (s *NotificationService) ProcessSaveChallan(payload map[string]interface{}) error {
	log.Printf("Notification consumer received save-challan event")
	return s.processChallanEvent(payload, "CREATED")
}

func (s *NotificationService) ProcessUpdateChallan(payload map[string]interface{}) error {
	log.Printf("Notification consumer received update-challan event")
	// Note: in Java it decides the action based on status (e.g. CANCELLED, PAID, UPDATED). We will just use UPDATED for simplicity or PAID if it's PAID.
	return s.processChallanEvent(payload, "UPDATED")
}

func (s *NotificationService) processChallanEvent(payload map[string]interface{}, defaultAction string) error {
	reqInfoMap, ok1 := payload["RequestInfo"].(map[string]interface{})
	challanMap, ok2 := payload["challan"].(map[string]interface{})
	if !ok1 || !ok2 {
		return fmt.Errorf("invalid payload format for notification")
	}

	reqInfoBytes, _ := json.Marshal(reqInfoMap)
	challanBytes, _ := json.Marshal(challanMap)

	var reqInfo domain.RequestInfo
	var challan domain.Challan

	if err := json.Unmarshal(reqInfoBytes, &reqInfo); err != nil { zap.L().Error("Failed to unmarshal RequestInfo", zap.Error(err)); return err }
	if err := json.Unmarshal(challanBytes, &challan); err != nil { zap.L().Error("Failed to unmarshal Challan", zap.Error(err)); return err }

	action := defaultAction
	if challan.ApplicationStatus == "PAID" {
		action = "PAID"
	} else if challan.ApplicationStatus == "CANCELLED" {
		action = "CANCELLED"
	}

	s.SendNotifications(&reqInfo, &challan, action)
	return nil
}
