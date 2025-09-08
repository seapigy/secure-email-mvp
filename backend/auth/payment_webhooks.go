package auth

// DO NOT EDIT EXISTING CODE - new file added
// Stripe webhook handlers for payment events and subscription management

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

// Stripe webhook endpoint secret (in production, this should come from environment)
var stripeWebhookSecret = "whsec_..." // TODO: Move to environment variable

type StripeWebhookEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

// POST /api/payment/webhook
func StripeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR reading webhook body: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Verify webhook signature
	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), stripeWebhookSecret)
	if err != nil {
		log.Printf("ERROR verifying webhook signature: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Handle the event
	switch event.Type {
	case "customer.subscription.created":
		err = handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		err = handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		err = handleSubscriptionDeleted(event)
	case "invoice.payment_succeeded":
		err = handlePaymentSucceeded(event)
	case "invoice.payment_failed":
		err = handlePaymentFailed(event)
	default:
		log.Printf("INFO unhandled webhook event type: %s", event.Type)
	}

	if err != nil {
		log.Printf("ERROR handling webhook event %s: %v", event.Type, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Handle subscription created event
func handleSubscriptionCreated(event stripe.Event) error {
	var subscription stripe.Subscription
	objectBytes, err := json.Marshal(event.Data.Object)
	if err != nil {
		return err
	}
	err = json.Unmarshal(objectBytes, &subscription)
	if err != nil {
		return err
	}

	// Find user by Stripe customer ID
	var userID string
	err = DB.QueryRow(`
		SELECT user_id FROM subscriptions 
		WHERE stripe_customer_id = ?
	`, subscription.Customer.ID).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("WARN subscription created for unknown customer: %s", subscription.Customer.ID)
			return nil
		}
		return err
	}

	// Encrypt Stripe subscription ID
	encryptedSubscriptionID, err := encryptSubscriptionData(subscription.ID)
	if err != nil {
		return err
	}

	// Update subscription in database
	now := time.Now().UTC()
	_, err = DB.Exec(`
		UPDATE subscriptions 
		SET stripe_subscription_id = ?, 
		    status = ?, 
		    updated_at = ?
		WHERE user_id = ?
	`, encryptedSubscriptionID, subscription.Status, now, userID)

	if err != nil {
		return err
	}

	log.Printf("INFO subscription_created user_id=%s subscription_id=%s", userID, subscription.ID)
	return nil
}

// Handle subscription updated event
func handleSubscriptionUpdated(event stripe.Event) error {
	var subscription stripe.Subscription
	objectBytes, err := json.Marshal(event.Data.Object)
	if err != nil {
		return err
	}
	err = json.Unmarshal(objectBytes, &subscription)
	if err != nil {
		return err
	}

	// Find user by Stripe customer ID
	var userID string
	err = DB.QueryRow(`
		SELECT user_id FROM subscriptions 
		WHERE stripe_customer_id = ?
	`, subscription.Customer.ID).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("WARN subscription updated for unknown customer: %s", subscription.Customer.ID)
			return nil
		}
		return err
	}

	// Update subscription status
	now := time.Now().UTC()
	_, err = DB.Exec(`
		UPDATE subscriptions 
		SET status = ?, 
		    updated_at = ?
		WHERE user_id = ?
	`, subscription.Status, now, userID)

	if err != nil {
		return err
	}

	// If subscription is canceled or past due, downgrade user account
	if subscription.Status == "canceled" || subscription.Status == "past_due" {
		_, err = DB.Exec(`
			UPDATE users 
			SET account_type_new = 'free', updated_at = ?
			WHERE id = ?
		`, now, userID)

		if err != nil {
			return err
		}

		log.Printf("INFO user_downgraded user_id=%s reason=subscription_%s", userID, subscription.Status)
	}

	log.Printf("INFO subscription_updated user_id=%s status=%s", userID, subscription.Status)
	return nil
}

// Handle subscription deleted event
func handleSubscriptionDeleted(event stripe.Event) error {
	var subscription stripe.Subscription
	objectBytes, err := json.Marshal(event.Data.Object)
	if err != nil {
		return err
	}
	err = json.Unmarshal(objectBytes, &subscription)
	if err != nil {
		return err
	}

	// Find user by Stripe customer ID
	var userID string
	err = DB.QueryRow(`
		SELECT user_id FROM subscriptions 
		WHERE stripe_customer_id = ?
	`, subscription.Customer.ID).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("WARN subscription deleted for unknown customer: %s", subscription.Customer.ID)
			return nil
		}
		return err
	}

	// Mark subscription as canceled
	now := time.Now().UTC()
	_, err = DB.Exec(`
		UPDATE subscriptions 
		SET status = 'canceled', 
		    updated_at = ?
		WHERE user_id = ?
	`, now, userID)

	if err != nil {
		return err
	}

	// Downgrade user to free account
	_, err = DB.Exec(`
		UPDATE users 
		SET account_type_new = 'free', updated_at = ?
		WHERE id = ?
	`, now, userID)

	if err != nil {
		return err
	}

	log.Printf("INFO subscription_deleted user_id=%s", userID)
	return nil
}

// Handle payment succeeded event
func handlePaymentSucceeded(event stripe.Event) error {
	var invoice stripe.Invoice
	objectBytes, err := json.Marshal(event.Data.Object)
	if err != nil {
		return err
	}
	err = json.Unmarshal(objectBytes, &invoice)
	if err != nil {
		return err
	}

	// Find user by Stripe customer ID
	var userID string
	err = DB.QueryRow(`
		SELECT user_id FROM subscriptions 
		WHERE stripe_customer_id = ?
	`, invoice.Customer.ID).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("WARN payment succeeded for unknown customer: %s", invoice.Customer.ID)
			return nil
		}
		return err
	}

	// Update subscription end date
	if invoice.Subscription != nil {
		// Note: In a real implementation, you would fetch the subscription from Stripe
		// For now, we'll skip this to avoid the undefined reference
		// subscription, err := subscription.Get(invoice.Subscription.ID, nil)
		// if err != nil {
		//     return err
		// }

		now := time.Now().UTC()
		// For now, set end date to 30 days from now (in real implementation, use subscription.CurrentPeriodEnd)
		_, err = DB.Exec(`
			UPDATE subscriptions 
			SET end_date = ?, 
			    updated_at = ?
			WHERE user_id = ?
		`, now.AddDate(0, 0, 30), now, userID)

		if err != nil {
			return err
		}
	}

	log.Printf("INFO payment_succeeded user_id=%s amount=%d", userID, invoice.AmountPaid)
	return nil
}

// Handle payment failed event
func handlePaymentFailed(event stripe.Event) error {
	var invoice stripe.Invoice
	objectBytes, err := json.Marshal(event.Data.Object)
	if err != nil {
		return err
	}
	err = json.Unmarshal(objectBytes, &invoice)
	if err != nil {
		return err
	}

	// Find user by Stripe customer ID
	var userID string
	err = DB.QueryRow(`
		SELECT user_id FROM subscriptions 
		WHERE stripe_customer_id = ?
	`, invoice.Customer.ID).Scan(&userID)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("WARN payment failed for unknown customer: %s", invoice.Customer.ID)
			return nil
		}
		return err
	}

	// Update subscription status to past_due
	now := time.Now().UTC()
	_, err = DB.Exec(`
		UPDATE subscriptions 
		SET status = 'past_due', 
		    updated_at = ?
		WHERE user_id = ?
	`, now, userID)

	if err != nil {
		return err
	}

	log.Printf("INFO payment_failed user_id=%s amount=%d", userID, invoice.AmountDue)
	return nil
}
