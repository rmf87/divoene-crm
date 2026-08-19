package commands

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/rmf87/divoene/internal/core/services"
	"github.com/rmf87/divoene/internal/infra/clients"
	"github.com/rmf87/divoene/internal/infra/database"
	"github.com/spf13/cobra"
)

var whatsappCmd = &cobra.Command{
	Use:   "whatsapp",
	Short: "WhatsApp helpers (dev/testing)",
}

// whatsappSendCmd sends an outbound message to a lead (mock mode without token).
var whatsappSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a WhatsApp message to a lead",
	RunE: func(cmd *cobra.Command, args []string) error {
		leadID, _ := cmd.Flags().GetString("lead")
		text, _ := cmd.Flags().GetString("text")
		if leadID == "" || text == "" {
			return fmt.Errorf("--lead and --text are required")
		}

		db, chatSvc, err := newChatService()
		if err != nil {
			return err
		}
		defer db.Close()

		msg, err := chatSvc.SendMessage(cmd.Context(), leadID, text)
		if err != nil {
			return err
		}
		fmt.Printf("Enviada: msg_id=%d wa_message_id=%s\n", msg.ID, msg.WAMessageID)
		return nil
	},
}

// whatsappSimulateCmd injects an inbound message as if the lead replied.
var whatsappSimulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Simulate an inbound message from a lead (dev/testing)",
	RunE: func(cmd *cobra.Command, args []string) error {
		from, _ := cmd.Flags().GetString("from")
		text, _ := cmd.Flags().GetString("text")
		if from == "" || text == "" {
			return fmt.Errorf("--from and --text are required")
		}

		db, chatSvc, err := newChatService()
		if err != nil {
			return err
		}
		defer db.Close()

		payload := services.MockInboundPayload(from, text)
		if err := chatSvc.HandleInbound(cmd.Context(), payload); err != nil {
			return err
		}
		fmt.Printf("Simulada mensagem de %s: %s\n", from, text)
		return nil
	},
}

func newChatService() (*sql.DB, *services.ChatService, error) {
	db, err := database.NewDB(resolveDBPath())
	if err != nil {
		return nil, nil, err
	}
	wa := clients.NewWhatsAppClient(
		os.Getenv("WHATSAPP_BASE_URL"), os.Getenv("WHATSAPP_TOKEN"),
		os.Getenv("WHATSAPP_PHONE_NUMBER_ID"), os.Getenv("WHATSAPP_APP_SECRET"),
		os.Getenv("WHATSAPP_WEBHOOK_VERIFY_TOKEN"))
	svc := services.NewChatService(database.NewChatMessageRepository(db), database.NewLeadRepository(db), wa, false)
	return db, svc, nil
}

func init() {
	whatsappSendCmd.Flags().String("lead", "", "lead id")
	whatsappSendCmd.Flags().String("text", "", "message text")
	whatsappSendCmd.MarkFlagRequired("lead")
	whatsappSendCmd.MarkFlagRequired("text")

	whatsappSimulateCmd.Flags().String("from", "", "lead phone (E.164, e.g. 5511999990001)")
	whatsappSimulateCmd.Flags().String("text", "", "message text")
	whatsappSimulateCmd.MarkFlagRequired("from")
	whatsappSimulateCmd.MarkFlagRequired("text")

	whatsappCmd.AddCommand(whatsappSendCmd)
	whatsappCmd.AddCommand(whatsappSimulateCmd)
	rootCmd.AddCommand(whatsappCmd)
}
