package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RegisterRoutes registers all sales API routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Apply common middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// API version group
	r.Route("/api/v1/sales", func(r chi.Router) {
		// Apply authentication middleware to all routes
		r.Use(h.AuthMiddleware)
		r.Use(h.TenantMiddleware)

		// Lead routes
		r.Route("/leads", func(r chi.Router) {
			r.Post("/", h.CreateLead)
			r.Get("/", h.ListLeads)

			// Statistics and reports
			r.Get("/statistics", h.GetLeadStatistics)
			r.Get("/by-owner/{ownerID}", h.GetLeadsByOwner)
			r.Get("/high-score", h.GetHighScoreLeads)
			r.Get("/unassigned", h.GetUnassignedLeads)
			r.Get("/stale", h.GetStaleLeads)

			// Bulk operations
			r.Post("/bulk/assign", h.BulkAssignLeads)
			r.Post("/bulk/status", h.BulkUpdateLeadStatus)
			r.Delete("/bulk", h.BulkDeleteLeads)

			// Single lead operations
			r.Route("/{leadID}", func(r chi.Router) {
				r.Get("/", h.GetLead)
				r.Put("/", h.UpdateLead)
				r.Delete("/", h.DeleteLead)

				// Status transitions
				r.Post("/qualify", h.QualifyLead)
				r.Post("/disqualify", h.DisqualifyLead)
				r.Post("/convert", h.ConvertLead)
				r.Post("/reactivate", h.ReactivateLead)

				// Assignment
				r.Post("/assign", h.AssignLead)
				r.Delete("/assign", h.UnassignLead)
			})
		})

		// Opportunity routes
		r.Route("/opportunities", func(r chi.Router) {
			r.Post("/", h.CreateOpportunity)
			r.Get("/", h.ListOpportunities)

			// Statistics and reports
			r.Get("/statistics", h.GetOpportunityStatistics)
			r.Get("/pipeline-value", h.GetPipelineValue)
			r.Get("/closing-this-month", h.GetClosingThisMonth)
			r.Get("/overdue", h.GetOverdueOpportunities)

			// Bulk operations
			r.Post("/bulk/assign", h.BulkAssignOpportunities)
			r.Post("/bulk/move-stage", h.BulkMoveStage)

			// Single opportunity operations
			r.Route("/{opportunityID}", func(r chi.Router) {
				r.Get("/", h.GetOpportunity)
				r.Put("/", h.UpdateOpportunity)
				r.Delete("/", h.DeleteOpportunity)

				// Stage transitions
				r.Post("/move-stage", h.MoveOpportunityToStage)
				r.Post("/win", h.WinOpportunity)
				r.Post("/lose", h.LoseOpportunity)
				r.Post("/reopen", h.ReopenOpportunity)
				r.Get("/stage-history", h.GetOpportunityStageHistory)

				// Products
				r.Route("/products", func(r chi.Router) {
					r.Post("/", h.AddOpportunityProduct)
					r.Put("/{productID}", h.UpdateOpportunityProduct)
					r.Delete("/{productID}", h.RemoveOpportunityProduct)
				})

				// Contacts
				r.Route("/contacts", func(r chi.Router) {
					r.Post("/", h.AddOpportunityContact)
					r.Put("/{contactID}", h.UpdateOpportunityContact)
					r.Delete("/{contactID}", h.RemoveOpportunityContact)
					r.Post("/{contactID}/set-primary", h.SetOpportunityPrimaryContact)
				})
			})
		})

		// Deal routes
		r.Route("/deals", func(r chi.Router) {
			r.Post("/", h.CreateDeal)
			r.Get("/", h.ListDeals)

			// Statistics and reports
			r.Get("/statistics", h.GetDealStatistics)
			r.Get("/by-owner/{ownerID}", h.GetDealsByOwner)
			r.Get("/by-customer/{customerID}", h.GetDealsByCustomer)
			r.Get("/overdue-invoices", h.GetOverdueInvoices)
			r.Get("/pending-payments", h.GetPendingPayments)
			r.Get("/revenue", h.GetRevenueByPeriod)

			// Bulk operations
			r.Post("/bulk/assign", h.BulkAssignDeals)
			r.Post("/bulk/status", h.BulkUpdateDealStatus)

			// Single deal operations
			r.Route("/{dealID}", func(r chi.Router) {
				r.Get("/", h.GetDeal)
				r.Put("/", h.UpdateDeal)
				r.Delete("/", h.DeleteDeal)

				// Status transitions
				r.Post("/submit", h.SubmitDealForApproval)
				r.Post("/approve", h.ApproveDeal)
				r.Post("/reject", h.RejectDeal)
				r.Post("/start-fulfillment", h.StartDealFulfillment)
				r.Post("/win", h.WinDeal)
				r.Post("/lose", h.LoseDeal)
				r.Post("/cancel", h.CancelDeal)
				r.Post("/reopen", h.ReopenDeal)

				// Line items
				r.Route("/line-items", func(r chi.Router) {
					r.Post("/", h.AddDealLineItem)
					r.Put("/{lineItemID}", h.UpdateDealLineItem)
					r.Delete("/{lineItemID}", h.RemoveDealLineItem)
				})

				// Invoices
				r.Route("/invoices", func(r chi.Router) {
					r.Post("/", h.CreateDealInvoice)
					r.Put("/{invoiceID}", h.UpdateDealInvoice)
					r.Post("/{invoiceID}/issue", h.IssueInvoice)
					r.Post("/{invoiceID}/cancel", h.CancelInvoice)
				})

				// Payments
				r.Route("/payments", func(r chi.Router) {
					r.Post("/", h.RecordPayment)
					r.Put("/{paymentID}", h.UpdatePayment)
					r.Post("/{paymentID}/refund", h.RefundPayment)
				})
			})
		})

		// Pipeline routes
		r.Route("/pipelines", func(r chi.Router) {
			r.Post("/", h.CreatePipeline)
			r.Get("/", h.ListPipelines)
			r.Get("/default", h.GetDefaultPipeline)
			r.Get("/by-type/{type}", h.GetPipelinesByType)

			// Single pipeline operations
			r.Route("/{pipelineID}", func(r chi.Router) {
				r.Get("/", h.GetPipeline)
				r.Put("/", h.UpdatePipeline)
				r.Delete("/", h.DeletePipeline)

				// Status operations
				r.Post("/activate", h.ActivatePipeline)
				r.Post("/deactivate", h.DeactivatePipeline)
				r.Post("/set-default", h.SetDefaultPipeline)
				r.Post("/clone", h.ClonePipeline)

				// Statistics and reports
				r.Get("/statistics", h.GetPipelineStatistics)
				r.Get("/velocity", h.GetPipelineVelocity)
				r.Get("/conversion-rates", h.GetStageConversionRates)
				r.Get("/forecast", h.GetForecast)

				// Stage operations
				r.Route("/stages", func(r chi.Router) {
					r.Post("/", h.AddPipelineStage)
					r.Put("/reorder", h.ReorderPipelineStages)

					r.Route("/{stageID}", func(r chi.Router) {
						r.Put("/", h.UpdatePipelineStage)
						r.Delete("/", h.RemovePipelineStage)
						r.Post("/activate", h.ActivatePipelineStage)
						r.Post("/deactivate", h.DeactivatePipelineStage)
					})
				})
			})
		})
	})
}

// NewRouter creates a new chi router with all sales routes registered
func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}
