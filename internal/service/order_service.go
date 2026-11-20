package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ahyan-1986/ag-ecommerce-backend/internal/config"
	"github.com/ahyan-1986/ag-ecommerce-backend/internal/domain"
	"github.com/ahyan-1986/ag-ecommerce-backend/pkg/pdf"
	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v82"
	checkoutsession "github.com/stripe/stripe-go/v82/checkout/session"
)

type orderService struct {
	orderRepo     domain.OrderRepository
	productRepo   domain.ProductRepository
	userRepo      domain.UserRepository
	voucherRepo   domain.VoucherRepository
	inventoryRepo domain.InventoryRepository
	affiliateSvc  domain.AffiliateService
	walletSvc     domain.WalletService
	emailSvc      domain.EmailService
	shippingSvc   domain.ShippingService
	cfg           *config.Config
}

func NewOrderService(
	orderRepo domain.OrderRepository,
	productRepo domain.ProductRepository,
	userRepo domain.UserRepository,
	voucherRepo domain.VoucherRepository,
	inventoryRepo domain.InventoryRepository,
	affiliateSvc domain.AffiliateService,
	walletSvc domain.WalletService,
	emailSvc domain.EmailService,
	shippingSvc domain.ShippingService,
	cfg *config.Config,
) domain.OrderService {
	stripe.Key = cfg.Stripe.SecretKey
	return &orderService{
		orderRepo:     orderRepo,
		productRepo:   productRepo,
		userRepo:      userRepo,
		voucherRepo:   voucherRepo,
		inventoryRepo: inventoryRepo,
		affiliateSvc:  affiliateSvc,
		walletSvc:     walletSvc,
		emailSvc:      emailSvc,
		shippingSvc:   shippingSvc,
		cfg:           cfg,
	}
}

func (s *orderService) GetCart(userID uuid.UUID) ([]domain.CartItem, error) {
	return s.orderRepo.GetCartItems(userID)
}

func (s *orderService) AddToCart(userID uuid.UUID, req domain.CartItemRequest) ([]domain.CartItem, error) {
	variant, err := s.productRepo.FindVariantByID(req.VariantID)
	if err != nil {
		return nil, errors.New("product variant not found")
	}
	if !variant.IsActive {
		return nil, errors.New("product variant is not available")
	}
	if variant.AvailableStock() < req.Quantity {
		return nil, fmt.Errorf("insufficient stock: only %d available", variant.AvailableStock())
	}

	item := &domain.CartItem{
		UserID:    userID,
		VariantID: req.VariantID,
		Quantity:  req.Quantity,
	}
	if err := s.orderRepo.UpsertCartItem(item); err != nil {
		return nil, err
	}
	return s.orderRepo.GetCartItems(userID)
}

func (s *orderService) UpdateCartItem(userID uuid.UUID, itemID uuid.UUID, qty int) ([]domain.CartItem, error) {
	if qty <= 0 {
		if err := s.orderRepo.DeleteCartItem(itemID, userID); err != nil {
			return nil, err
		}
		return s.orderRepo.GetCartItems(userID)
	}
	items, err := s.orderRepo.GetCartItems(userID)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.ID == itemID {
			item.Quantity = qty
			if err := s.orderRepo.UpsertCartItem(&item); err != nil {
				return nil, err
			}
			break
		}
	}
	return s.orderRepo.GetCartItems(userID)
}

func (s *orderService) RemoveCartItem(userID uuid.UUID, itemID uuid.UUID) error {
	return s.orderRepo.DeleteCartItem(itemID, userID)
}

func (s *orderService) CaptureStripePayment(userID uuid.UUID, orderNo, sessionID string) (*domain.Order, error) {
	order, err := s.orderRepo.FindByOrderNo(orderNo)
	if err != nil {
		return nil, errors.New("order not found")
	}

	user, _ := s.userRepo.FindByID(userID)
	if user != nil && !user.IsAdmin() && order.UserID != userID {
		return nil, errors.New("access denied")
	}

	// Already confirmed (idempotent)
	if order.Status == domain.OrderStatusPaid {
		return order, nil
	}

	// Prefer session ID passed directly from Stripe redirect over stored ref
	ref := sessionID
	if ref == "" {
		ref = order.PaymentRef
	}
	if ref == "" {
		return nil, errors.New("payment session missing")
	}

	// Persist payment ref if not yet stored
	if order.PaymentRef == "" {
		order.PaymentRef = ref
		s.orderRepo.Update(order) // best-effort
	}

	sess, err := checkoutsession.Get(ref, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe_error: %w", err)
	}

	if sess.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		return nil, fmt.Errorf("payment_pending: status is %s", sess.PaymentStatus)
	}

	now := time.Now()
	order.Status = domain.OrderStatusPaid
	order.PaidAt = &now
	order.PaymentRef = ref
	if err := s.orderRepo.Update(order); err != nil {
		return nil, err
	}

	// Send receipt asynchronously
	s.sendReceiptAsync(order)

	return order, nil
}

func (s *orderService) GetOrderByNo(userID uuid.UUID, orderNo string) (*domain.Order, error) {
	order, err := s.orderRepo.FindByOrderNo(orderNo)
	if err != nil {
		return nil, errors.New("order not found")
	}
	user, _ := s.userRepo.FindByID(userID)
	if user != nil && !user.IsAdmin() && order.UserID != userID {
		return nil, errors.New("access denied")
	}
	return order, nil
}

func (s *orderService) Checkout(userID uuid.UUID, req domain.CheckoutRequest) (*domain.Order, string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, "", errors.New("user not found")
	}

	cartItems, err := s.orderRepo.GetCartItems(userID)
	if err != nil || len(cartItems) == 0 {
		return nil, "", errors.New("cart is empty")
	}

	// Calculate prices and BV using role-based pricing
	var subTotal float64
	var totalBV float64
	orderItems := make([]domain.OrderItem, 0, len(cartItems))
	for _, ci := range cartItems {
		if ci.Variant == nil {
			continue
		}
		// Stock re-check
		if ci.Variant.AvailableStock() < ci.Quantity {
			return nil, "", fmt.Errorf("insufficient stock for %s", ci.Variant.Name)
		}
		unitPrice := ci.Variant.RetailPrice
		if user.IsDealer() && user.DealerApproved && ci.Variant.DealerPrice > 0 {
			unitPrice = ci.Variant.DealerPrice
		}
		lineTotal := unitPrice * float64(ci.Quantity)
		subTotal += lineTotal

		itemBV := ci.Variant.BV
		if itemBV <= 0 {
			itemBV = unitPrice
		}
		lineBV := itemBV * float64(ci.Quantity)
		totalBV += lineBV

		// Fetch product info for snapshot
		product, _ := s.productRepo.FindByID(ci.Variant.ProductID)
		productName := ""
		if product != nil {
			productName = product.Name
		}

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:   ci.Variant.ProductID,
			VariantID:   ci.VariantID,
			ProductName: productName,
			VariantName: ci.Variant.Name,
			SKU:         ci.Variant.SKU,
			ImageURL:    ci.Variant.ImageURL,
			Quantity:    ci.Quantity,
			UnitPrice:   unitPrice,
			TotalPrice:  lineTotal,
			BV:          lineBV,
		})
	}

	// Calculate shipping fee from provided rate
	shippingFee := 0.0
	var totalWeightKg float64
	var totalItems int
	for _, item := range orderItems {
		totalItems += item.Quantity
		// Attempt to parse weight if product variant has it, or we could fetch it.
		// For now we will approximate or check if variant has weight.
		// Since Variant might not have WeightGrams loaded directly in cartItems loop above,
		// let's fetch it from the database if needed, but since we already loaded it:
		// Oh, variant doesn't have WeightGrams? Let's check `domain.ProductVariant`.
		// Let's assume weight is not loaded or we do a simple fallback.
	}

	if req.ShippingRateID != "" {
		rateID, err := uuid.Parse(req.ShippingRateID)
		if err == nil {
			rate, _ := s.shippingSvc.GetRate(rateID)
			if rate != nil {
				// Re-calculate the actual weight/dimensions
				// We need to fetch product variants to get weight.
				for _, ci := range cartItems {
					if ci.Variant != nil {
						totalWeightKg += ci.Variant.Weight * float64(ci.Quantity)
					}
				}
				calcReq := domain.CalculateShippingRequest{
					Country:   req.ShippingCountry,
					State:     req.ShippingState,
					WeightKg:  totalWeightKg,
					OrderRM:   subTotal,
					ItemCount: totalItems,
				}
				shippingFee = CalcCost(*rate, calcReq)
			}
		}
	}
	var discountAmount float64
	var voucherID *uuid.UUID
	if req.VoucherCode != "" {
		vRes, err := s.validateAndApplyVoucher(userID, req.VoucherCode, subTotal)
		if err != nil {
			return nil, "", fmt.Errorf("voucher error: %w", err)
		}
		discountAmount = vRes.DiscountAmount
		voucherID = &vRes.Voucher.ID
	}

	totalBeforePoints := subTotal + shippingFee - discountAmount
	if totalBeforePoints < 0 {
		totalBeforePoints = 0
	}

	// Apply Loyalty Points at checkout (Aameba PointsBlock: 1 point = RM 1)
	var pointsUsed float64
	var pointsDiscount float64
	if req.PointsToUse > 0 && s.walletSvc != nil {
		wallet, err := s.walletSvc.GetOrCreateWallet(userID)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read loyalty points wallet: %w", err)
		}
		pts := req.PointsToUse
		if pts > wallet.PointsBalance {
			return nil, "", fmt.Errorf("insufficient loyalty points (available: %.2f)", wallet.PointsBalance)
		}
		if pts > totalBeforePoints {
			pts = totalBeforePoints
		}
		pointsUsed = round2(pts)
		pointsDiscount = pointsUsed
	}

	totalAmount := round2(totalBeforePoints - pointsDiscount)
	if totalAmount < 0 {
		totalAmount = 0
	}

	country := req.ShippingCountry
	if country == "" {
		country = "Malaysia"
	}

	orderNo := generateOrderNo()
	order := &domain.Order{
		OrderNo:          orderNo,
		UserID:           userID,
		BuyerType:        "M",
		SponsorCode:      user.ReferredByCode,
		Status:           domain.OrderStatusPending,
		SubTotal:         subTotal,
		ShippingFee:      shippingFee,
		DiscountAmount:   discountAmount,
		PointsUsed:       pointsUsed,
		PointsDiscount:   pointsDiscount,
		TotalAmount:      totalAmount,
		TotalBV:          totalBV,
		Currency:         "MYR",
		VoucherCode:      req.VoucherCode,
		VoucherID:        voucherID,
		PaymentMethod:    req.PaymentMethod,
		ShippingName:     req.ShippingName,
		ShippingPhone:    req.ShippingPhone,
		ShippingLine1:    req.ShippingLine1,
		ShippingLine2:    req.ShippingLine2,
		ShippingCity:     req.ShippingCity,
		ShippingState:    req.ShippingState,
		ShippingPostcode: req.ShippingPostcode,
		ShippingCountry:  country,
		ShippingStatus:   domain.ShippingNew,
		Notes:            req.Notes,
		Items:            orderItems,
	}

	// Handle COD, full-points payment, or demo instant confirmation
	if req.PaymentMethod == "cod" || req.PaymentMethod == "points" || totalAmount == 0 || s.cfg.Stripe.SecretKey == "" {
		order.Status = domain.OrderStatusPaid
		now := time.Now()
		order.PaidAt = &now
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, "", fmt.Errorf("create order: %w", err)
	}

	// Debit loyalty points if used
	if pointsUsed > 0 && s.walletSvc != nil {
		desc := fmt.Sprintf("Redeemed at Checkout · Order %s", order.OrderNo)
		if _, err := s.walletSvc.DebitPoints(userID, pointsUsed, "checkout_redeem", order.OrderNo, desc); err != nil {
			return nil, "", fmt.Errorf("failed to debit loyalty points: %w", err)
		}
	}

	// Reserve stock
	for _, item := range orderItems {
		v, err := s.productRepo.FindVariantByID(item.VariantID)
		if err != nil {
			continue
		}
		v.StockReserved += item.Quantity
		s.productRepo.UpdateVariant(v)
		// Record inventory movement
		s.inventoryRepo.CreateMovement(&domain.InventoryMovement{
			VariantID:     item.VariantID,
			Type:          domain.MovementSale,
			Quantity:      -item.Quantity,
			BalanceBefore: v.StockOnHand,
			BalanceAfter:  v.StockOnHand - item.Quantity,
			Reference:     order.OrderNo,
			Note:          "Sale order " + order.OrderNo,
			CreatedBy:     userID,
		})
		v.StockOnHand -= item.Quantity
		v.StockReserved -= item.Quantity
		s.productRepo.UpdateVariant(v)
	}

	// Apply voucher redemption record & mark claimed voucher used
	if req.VoucherCode != "" && voucherID != nil {
		s.voucherRepo.IncrementUsed(*voucherID)
		_ = s.voucherRepo.MarkUserVoucherUsed(*voucherID, userID, order.ID)
		s.voucherRepo.CreateRedemption(&domain.VoucherRedemption{
			VoucherID:      *voucherID,
			UserID:         userID,
			OrderID:        order.ID,
			OriginalAmount: subTotal,
			DiscountAmount: discountAmount,
			FinalAmount:    totalAmount,
		})
	}

	// Clear cart
	s.orderRepo.ClearCart(userID)

	// Process Aameba MLM Bonus Engine (Self RB + Upline RB + Differential LRB + Rank Promotion)
	if order.Status == domain.OrderStatusPaid && s.affiliateSvc != nil {
		_ = s.affiliateSvc.ProcessOrderCommission(order)
	}

	// Send order confirmation email
	if s.emailSvc != nil {
		go s.emailSvc.SendOrderConfirmation(user.Email, user.FirstName, order.OrderNo, order.TotalAmount)
	}

	// Create Stripe Checkout Session for online payment
	var paymentURL string
	if req.PaymentMethod == "stripe" && s.cfg.Stripe.SecretKey != "" {
		frontendURL := s.cfg.Email.AppURL
		if frontendURL == "" {
			frontendURL = "http://localhost:5173"
		}

		lineItems := []*stripe.CheckoutSessionLineItemParams{}
		for _, item := range orderItems {
			name := item.ProductName
			if item.VariantName != "" && item.VariantName != item.ProductName {
				name = item.ProductName + " — " + item.VariantName
			}
			lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(s.cfg.Stripe.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(name),
					},
					UnitAmount: stripe.Int64(int64(item.UnitPrice * 100)),
				},
				Quantity: stripe.Int64(int64(item.Quantity)),
			})
		}
		if shippingFee > 0 {
			lineItems = append(lineItems, &stripe.CheckoutSessionLineItemParams{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(s.cfg.Stripe.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Shipping"),
					},
					UnitAmount: stripe.Int64(int64(shippingFee * 100)),
				},
				Quantity: stripe.Int64(1),
			})
		}

		sess, err := checkoutsession.New(&stripe.CheckoutSessionParams{
			PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
			LineItems:          lineItems,
			Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
			SuccessURL:         stripe.String(frontendURL + "/orders?paid=1&order=" + order.OrderNo + "&session_id={CHECKOUT_SESSION_ID}"),
			CancelURL:          stripe.String(frontendURL + "/checkout"),
			CustomerEmail:      stripe.String(user.Email),
			Metadata: map[string]string{
				"order_id": order.ID.String(),
				"order_no": order.OrderNo,
			},
		})
		if err == nil {
			paymentURL = sess.URL
			order.PaymentRef = sess.ID
			s.orderRepo.Update(order)
		}
	}

	return order, paymentURL, nil
}

func (s *orderService) ConfirmStripePayment(sessionID string) error {
	// Called via Stripe webhook on checkout.session.completed
	// Find order by exact payment ref (Stripe Checkout Session ID)
	order, err := s.orderRepo.FindByPaymentRef(sessionID)
	if err != nil || order == nil {
		return errors.New("order not found for checkout session")
	}
	if order.Status == domain.OrderStatusPaid {
		return nil
	}
	now := time.Now()
	order.Status = domain.OrderStatusPaid
	order.PaidAt = &now
	if err := s.orderRepo.Update(order); err != nil {
		return err
	}

	// Reload order with full details for receipt generation
	fullOrder, reloadErr := s.orderRepo.FindByID(order.ID)
	if reloadErr != nil {
		fullOrder = order
	}

	// Send receipt asynchronously
	s.sendReceiptAsync(fullOrder)

	return nil
}

func (s *orderService) GetOrder(id uuid.UUID, userID uuid.UUID) (*domain.Order, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	// Customers can only see their own orders
	user, _ := s.userRepo.FindByID(userID)
	if user != nil && !user.IsAdmin() && order.UserID != userID {
		return nil, errors.New("access denied")
	}
	return order, nil
}

func (s *orderService) ListUserOrders(userID uuid.UUID, page, pageSize int) ([]domain.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return s.orderRepo.FindByUserID(userID, page, pageSize)
}

func (s *orderService) ListAllOrders(status domain.OrderStatus, search string, page, pageSize int) ([]domain.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.orderRepo.List(status, search, page, pageSize)
}

func (s *orderService) UpdateFulfillment(orderID uuid.UUID, req domain.UpdateFulfillmentRequest) (*domain.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	order.ShippingStatus = req.ShippingStatus
	if req.TrackingNumber != "" {
		order.TrackingNumber = req.TrackingNumber
	}
	if req.Courier != "" {
		order.Courier = req.Courier
	}
	if req.Notes != "" {
		order.Notes = req.Notes
	}
	// Sync order status with shipping status
	switch req.ShippingStatus {
	case domain.ShippingShipped:
		order.Status = domain.OrderStatusShipped
	case domain.ShippingDelivered:
		order.Status = domain.OrderStatusDelivered
	case domain.ShippingCancelled:
		order.Status = domain.OrderStatusCancelled
	}
	if err := s.orderRepo.Update(order); err != nil {
		return nil, err
	}
	// Notify customer: shipped
	if req.ShippingStatus == domain.ShippingShipped && s.emailSvc != nil {
		user, _ := s.userRepo.FindByID(order.UserID)
		if user != nil {
			go s.emailSvc.SendOrderShipped(user.Email, user.FirstName, order.OrderNo, order.Courier, order.TrackingNumber)
		}
	}
	return order, nil
}

func (s *orderService) CancelOrder(orderID uuid.UUID, userID uuid.UUID) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	user, _ := s.userRepo.FindByID(userID)
	if user != nil && !user.IsAdmin() && order.UserID != userID {
		return errors.New("access denied")
	}
	if order.Status == domain.OrderStatusDelivered || order.Status == domain.OrderStatusCancelled {
		return errors.New("order cannot be cancelled in its current state")
	}
	order.Status = domain.OrderStatusCancelled
	order.ShippingStatus = domain.ShippingCancelled
	return s.orderRepo.Update(order)
}

// sendReceiptAsync generates a PDF receipt and emails it to the customer.
func (s *orderService) sendReceiptAsync(order *domain.Order) {
	go func() {
		user := order.User
		if user == nil {
			u, err := s.userRepo.FindByID(order.UserID)
			if err != nil {
				log.Printf("[receipt] failed to find user for order %s: %v", order.OrderNo, err)
				return
			}
			user = u
		}

		pdfBytes, err := pdf.GenerateReceipt(order)
		if err != nil {
			log.Printf("[receipt] failed to generate PDF for order %s: %v", order.OrderNo, err)
			return
		}

		if s.emailSvc != nil {
			name := user.FirstName
			if name == "" {
				name = "Customer"
			}
			if err := s.emailSvc.SendReceiptWithPDF(user.Email, name, order.OrderNo, pdfBytes); err != nil {
				log.Printf("[receipt] failed to email receipt for order %s: %v", order.OrderNo, err)
			} else {
				log.Printf("[receipt] sent receipt email for order %s to %s", order.OrderNo, user.Email)
			}
		}
	}()
}

func (s *orderService) GetDashboardStats() (*domain.DashboardStats, error) {
	totalRevenue, _ := s.orderRepo.TotalRevenue()
	totalOrders, _ := s.orderRepo.CountByStatus("")
	pendingOrders, _ := s.orderRepo.CountByStatus(domain.OrderStatusPending)
	totalCustomers, _ := s.userRepo.CountByRole(domain.RoleCustomer)
	totalDealers, _ := s.userRepo.CountByRole(domain.RoleDealer)
	lowStock, _ := s.inventoryRepo.CountLowStock()
	monthlyRevenue, _ := s.orderRepo.RevenueByMonth(12)

	return &domain.DashboardStats{
		TotalRevenue:     totalRevenue,
		TotalOrders:      totalOrders,
		TotalCustomers:   totalCustomers,
		TotalDealers:     totalDealers,
		PendingOrders:    pendingOrders,
		LowStockProducts: lowStock,
		MonthlyRevenue:   monthlyRevenue,
	}, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (s *orderService) validateAndApplyVoucher(userID uuid.UUID, code string, cartTotal float64) (*domain.ValidateVoucherResponse, error) {
	v, err := s.voucherRepo.FindByCode(code)
	if err != nil {
		return nil, errors.New("voucher not found")
	}
	if !v.Active {
		return nil, errors.New("voucher is inactive")
	}
	now := time.Now()
	if v.StartsAt != nil && now.Before(*v.StartsAt) {
		return nil, errors.New("voucher is not yet valid")
	}
	if v.ExpiresAt != nil && now.After(*v.ExpiresAt) {
		return nil, errors.New("voucher has expired")
	}
	if v.MaxQty > 0 && v.UsedQty >= v.MaxQty {
		return nil, errors.New("voucher has been fully redeemed")
	}
	if cartTotal < v.MinSpend {
		return nil, fmt.Errorf("minimum spend of MYR %.2f required", v.MinSpend)
	}
	if v.PerUserLimit > 0 {
		used, _ := s.voucherRepo.CountRedemptionsByUser(v.ID, userID)
		if used >= v.PerUserLimit {
			return nil, errors.New("you have already used this voucher")
		}
	}
	discount := 0.0
	if v.Type == domain.VoucherTypePercent {
		discount = cartTotal * (v.Value / 100)
		if v.MaxDiscount > 0 && discount > v.MaxDiscount {
			discount = v.MaxDiscount
		}
	} else {
		discount = v.Value
		if discount > cartTotal {
			discount = cartTotal
		}
	}
	return &domain.ValidateVoucherResponse{
		Voucher:        *v,
		DiscountAmount: discount,
		FinalTotal:     cartTotal - discount,
	}, nil
}

func (s *orderService) GuestCheckout(req domain.GuestCheckoutRequest) (*domain.Order, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("cart is empty")
	}
	var subTotal float64
	var totalBV float64
	orderItems := make([]domain.OrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		v, err := s.productRepo.FindVariantByID(it.VariantID)
		if err != nil {
			return nil, errors.New("product variant not found")
		}
		if v.AvailableStock() < it.Quantity {
			return nil, fmt.Errorf("insufficient stock for %s", v.Name)
		}
		lineTotal := v.RetailPrice * float64(it.Quantity)
		subTotal += lineTotal
		bv := v.BV
		if bv <= 0 {
			bv = v.RetailPrice
		}
		lineBV := bv * float64(it.Quantity)
		totalBV += lineBV
		prod, _ := s.productRepo.FindByID(v.ProductID)
		prodName := ""
		if prod != nil {
			prodName = prod.Name
		}
		orderItems = append(orderItems, domain.OrderItem{
			ProductID:   v.ProductID,
			VariantID:   v.ID,
			ProductName: prodName,
			VariantName: v.Name,
			SKU:         v.SKU,
			ImageURL:    v.ImageURL,
			Quantity:    it.Quantity,
			UnitPrice:   v.RetailPrice,
			TotalPrice:  lineTotal,
			BV:          lineBV,
		})
	}

	country := req.ShippingCountry
	if country == "" {
		country = "Malaysia"
	}
	// Aameba GuestRebate: freeze 15% self-rebate on guest_points_held until guest registers
	guestHeld := round2(totalBV * 0.15)
	now := time.Now()
	systemGuestOwner := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	order := &domain.Order{
		OrderNo:          generateOrderNo(),
		UserID:           systemGuestOwner,
		BuyerType:        "G",
		GuestEmail:       req.GuestEmail,
		GuestName:        req.GuestName,
		GuestPhone:       req.GuestPhone,
		SponsorCode:      req.SponsorCode,
		Status:           domain.OrderStatusPaid,
		PaidAt:           &now,
		SubTotal:         subTotal,
		TotalAmount:      subTotal,
		TotalBV:          totalBV,
		GuestPointsHeld:  guestHeld,
		Currency:         "MYR",
		PaymentMethod:    req.PaymentMethod,
		ShippingName:     req.GuestName,
		ShippingPhone:    req.GuestPhone,
		ShippingLine1:    req.ShippingLine1,
		ShippingLine2:    req.ShippingLine2,
		ShippingCity:     req.ShippingCity,
		ShippingState:    req.ShippingState,
		ShippingPostcode: req.ShippingPostcode,
		ShippingCountry:  country,
		ShippingStatus:   domain.ShippingNew,
		Notes:            req.Notes,
		Items:            orderItems,
	}
	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("create guest order: %w", err)
	}
	if s.affiliateSvc != nil {
		_ = s.affiliateSvc.ProcessOrderCommission(order)
	}
	return order, nil
}

func generateOrderNo() string {
	now := time.Now()
	id := uuid.New().String()
	return fmt.Sprintf("AG-%d%02d-%s", now.Year(), now.Month(), id[:6])
}
