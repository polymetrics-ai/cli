# pm connectors inspect xero

```text
NAME
  pm connectors inspect xero - Xero connector manual

SYNOPSIS
  pm connectors inspect xero
  pm connectors inspect xero --json
  pm credentials add <name> --connector xero [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Xero Accounting API resources through declarative JSON bundle streams and typed write actions.

ICON
  id: xero
  asset: icons/xero.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.xero.com/documentation/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id
  bank_transaction_id
  bank_transfer_id
  base_url
  batch_payment_id
  branding_theme_id
  budget_id
  contact_group_id
  contact_id
  contact_number
  credit_note_id
  expense_claim_id
  invoice_id
  item_id
  journal_id
  journal_number
  linked_transaction_id
  manual_journal_id
  organisation_id
  overpayment_id
  payment_id
  prepayment_id
  purchase_order_id
  purchase_order_number
  quote_id
  receipt_id
  repeating_invoice_id
  report_id
  tax_type
  tracking_category_id
  user_id
  access_token (secret) (required)
  tenant_id (secret) (required)

ETL STREAMS
  invoices:
    primary key: InvoiceID
    cursor: UpdatedDateUTC
    fields: AmountDue(number), AmountPaid(number), ContactID(string), CurrencyCode(string), Date(string), DueDate(string), InvoiceID(string), InvoiceNumber(string), LineAmountTypes(string), Status(string), SubTotal(number), Total(number), TotalTax(number), Type(string), UpdatedDateUTC(string), id(string), status(string), type(string), updated_at(string)
  contacts:
    primary key: ContactID
    cursor: UpdatedDateUTC
    fields: AccountNumber(string), ContactID(string), ContactStatus(string), EmailAddress(string), FirstName(string), IsCustomer(boolean), IsSupplier(boolean), LastName(string), Name(string), UpdatedDateUTC(string), id(string), status(string), updated_at(string)
  accounts:
    primary key: AccountID
    cursor: UpdatedDateUTC
    fields: AccountID(string), Class(string), Code(string), EnablePaymentsToAccount(boolean), Name(string), Status(string), TaxType(string), Type(string), UpdatedDateUTC(string), id(string), status(string), type(string), updated_at(string)
  bank_transactions:
    primary key: BankTransactionID
    cursor: UpdatedDateUTC
    fields: BankTransactionID(string), ContactID(string), CurrencyCode(string), Date(string), IsReconciled(boolean), Status(string), SubTotal(number), Total(number), TotalTax(number), Type(string), UpdatedDateUTC(string), id(string), status(string), type(string), updated_at(string)
  items:
    primary key: ItemID
    cursor: UpdatedDateUTC
    fields: Code(string), Description(string), IsPurchased(boolean), IsSold(boolean), IsTrackedAsInventory(boolean), ItemID(string), Name(string), QuantityOnHand(number), UpdatedDateUTC(string), id(string), updated_at(string)
  payments:
    primary key: PaymentID
    cursor: UpdatedDateUTC
    fields: Amount(number), CurrencyRate(number), Date(string), PaymentID(string), PaymentType(string), Reference(string), Status(string), UpdatedDateUTC(string), id(string), status(string), updated_at(string)
  account:
    primary key: AccountID
    fields: AccountID(string), AddToWatchlist(boolean), BankAccountNumber(string), BankAccountType(string), Class(string), Code(string), CurrencyCode(string), Description(string), EnablePaymentsToAccount(boolean), HasAttachments(boolean), Name(string), ReportingCode(string), ReportingCodeName(string), ShowInExpenseClaims(boolean), Status(string), SystemAccount(string), TaxType(string), Type(string), UpdatedDateUTC(string), ValidationErrors(array), account_id(string)
  account_attachments:
    primary key: AttachmentID
    fields: AttachmentID(string), ContentLength(integer), FileName(string), IncludeOnline(boolean), MimeType(string), Url(string), account_id(string)
  bank_transaction:
    primary key: BankTransactionID
    fields: BankAccount(object), BankTransactionID(string), Contact(object), CurrencyCode(string), CurrencyRate(number), Date(string), HasAttachments(boolean), IsReconciled(boolean), LineAmountTypes(string), LineItems(array), OverpaymentID(string), PrepaymentID(string), Reference(string), Status(string), StatusAttributeString(string), SubTotal(number), Total(number), TotalTax(number), Type(string), UpdatedDateUTC(string), Url(string), ValidationErrors(array), bank_transaction_id(string)
  bank_transaction_attachments:
    primary key: AttachmentID
    fields: AttachmentID(string), ContentLength(integer), FileName(string), IncludeOnline(boolean), MimeType(string), Url(string), bank_transaction_id(string)
  bank_transactions_history:
    primary key: DateUTC
    fields: DateUTC(string)
  bank_transfer:
    primary key: BankTransferID
    fields: Amount(number), BankTransferID(string), CreatedDateUTC(string), CurrencyRate(number), Date(string), FromBankAccount(object), FromBankTransactionID(string), FromIsReconciled(boolean), HasAttachments(boolean), Reference(string), ToBankAccount(object), ToBankTransactionID(string), ToIsReconciled(boolean), ValidationErrors(array), bank_transfer_id(string)
  bank_transfer_attachments:
    primary key: AttachmentID
    fields: AttachmentID(string), ContentLength(integer), FileName(string), IncludeOnline(boolean), MimeType(string), Url(string), bank_transfer_id(string)
  bank_transfer_history:
    primary key: DateUTC
    fields: DateUTC(string)
  bank_transfers:
    primary key: BankTransferID
    fields: Amount(number), BankTransferID(string), CreatedDateUTC(string), CurrencyRate(number), Date(string), FromBankAccount(object), FromBankTransactionID(string), FromIsReconciled(boolean), HasAttachments(boolean), Reference(string), ToBankAccount(object), ToBankTransactionID(string), ToIsReconciled(boolean), ValidationErrors(array)
  batch_payment:
    primary key: BatchPaymentID
    fields: Account(object), Amount(number), BatchPaymentID(string), Code(string), Date(string), DateString(string), Details(string), IsReconciled(boolean), Narrative(string), Particulars(string), Payments(array), Reference(string), Status(string), TotalAmount(number), Type(string), UpdatedDateUTC(string), ValidationErrors(array), batch_payment_id(string)
  batch_payment_history:
    primary key: DateUTC
    fields: Changes(string), DateUTC(string), Details(string), User(string)
  batch_payments:
    primary key: BatchPaymentID
    fields: Account(object), Amount(number), BatchPaymentID(string), Code(string), Date(string), DateString(string), Details(string), IsReconciled(boolean), Narrative(string), Particulars(string), Payments(array), Reference(string), Status(string), TotalAmount(number), Type(string), UpdatedDateUTC(string), ValidationErrors(array)
  branding_theme:
    primary key: BrandingThemeID
    fields: BrandingThemeID(string), CreatedDateUTC(string), LogoUrl(string), Name(string), SortOrder(integer), Type(string), branding_theme_id(string)
  branding_theme_payment_services:
    primary key: PaymentServiceID
    fields: PayNowText(string), PaymentServiceID(string), PaymentServiceName(string), PaymentServiceType(string), PaymentServiceUrl(string), ValidationErrors(array), branding_theme_id(string)
  branding_themes:
    primary key: BrandingThemeID
    fields: BrandingThemeID(string), CreatedDateUTC(string), LogoUrl(string), Name(string), SortOrder(integer), Type(string)
  budget:
    primary key: BudgetID
    fields: BudgetID(string), BudgetLines(array), Description(string), Tracking(array), Type(string), UpdatedDateUTC(string), budget_id(string)
  budgets:
    primary key: BudgetID
    fields: BudgetID(string), BudgetLines(array), Description(string), Tracking(array), Type(string), UpdatedDateUTC(string)
  contact:
    primary key: ContactID
    fields: AccountNumber(string), AccountsPayableTaxType(string), AccountsReceivableTaxType(string), Addresses(array), Attachments(array), Balances(object), BankAccountDetails(string), BatchPayments(object), BrandingTheme(object), CompanyNumber(string), ContactGroups(array), ContactID(string), ContactNumber(string), ContactPersons(array), ContactStatus(string), DefaultCurrency(string), Discount(number), EmailAddress(string), FirstName(string), HasAttachments(boolean), HasValidationErrors(boolean), IsCustomer(boolean), IsSupplier(boolean), LastName(string), MergedToContactID(string), Name(string), PaymentTerms(object), Phones(array), PurchasesDefaultAccountCode(string), PurchasesDefaultLineAmountType(string), PurchasesTrackingCategories(array), SalesDefaultAccountCode(string), SalesDefaultLineAmountType(string), SalesTrackingCategories(array), StatusAttributeString(string), TaxNumber(string), TaxNumberType(string), TrackingCategoryName(string), TrackingCategoryOption(string), UpdatedDateUTC(string), ValidationErrors(array), Website(string), XeroNetworkKey(string), contact_id(string)
  contact_attachments:
    primary key: AttachmentID
    fields: AttachmentID(string), ContentLength(integer), FileName(string), IncludeOnline(boolean), MimeType(string), Url(string), contact_id(string)
  contact_by_contact_number:
    primary key: ContactID
    fields: AccountNumber(string), AccountsPayableTaxType(string), AccountsReceivableTaxType(string), Addresses(array), Attachments(array), Balances(object), BankAccountDetails(string), BatchPayments(object), BrandingTheme(object), CompanyNumber(string), ContactGroups(array), ContactID(string), ContactNumber(string), ContactPersons(array), ContactStatus(string), DefaultCurrency(string), Discount(number), EmailAddress(string), FirstName(string), HasAttachments(boolean), HasValidationErrors(boolean), IsCustomer(boolean), IsSupplier(boolean), LastName(string), MergedToContactID(string), Name(string), PaymentTerms(object), Phones(array), PurchasesDefaultAccountCode(string), PurchasesDefaultLineAmountType(string), PurchasesTrackingCategories(array), SalesDefaultAccountCode(string), SalesDefaultLineAmountType(string), SalesTrackingCategories(array), StatusAttributeString(string), TaxNumber(string), TaxNumberType(string), TrackingCategoryName(string), TrackingCategoryOption(string), UpdatedDateUTC(string), ValidationErrors(array), Website(string), XeroNetworkKey(string), contact_number(string)
  contact_cis_settings:
    primary key: contact_id
    fields: CISEnabled(boolean), Rate(number), contact_id(string)
  contact_group:
    primary key: ContactGroupID
    fields: ContactGroupID(string), Contacts(array), Name(string), Status(string), contact_group_id(string)
  contact_groups:
    primary key: ContactGroupID
    fields: ContactGroupID(string), Contacts(array), Name(string), Status(string)
  contact_history:
    primary key: DateUTC
    fields: DateUTC(string)
  credit_note:
    primary key: CreditNoteID
    fields: Allocations(array), AppliedAmount(number), BrandingThemeID(string), CISDeduction(number), CISRate(number), Contact(object), CreditNoteID(string), CreditNoteNumber(string), CurrencyCode(string), CurrencyRate(number), Date(string), DueDate(string), FullyPaidOnDate(string), HasAttachments(boolean), HasErrors(boolean), InvoiceAddresses(array), LineAmountTypes(string), LineItems(array), Payments(array), Reference(string), RemainingCredit(number), SentToContact(boolean), Status(string), StatusAttributeString(string), SubTotal(number), Total(number), TotalTax(number), Type(string), UpdatedDateUTC(string), ValidationErrors(array), Warnings(array), credit_note_id(string)
  credit_note_attachments:
    primary key: AttachmentID
    fields: AttachmentID(string), ContentLength(integer), FileName(string), IncludeOnline(boolean), MimeType(string), Url(string), credit_note_id(string)
  credit_note_history:
    primary key: DateUTC
    fields: DateUTC(string)
  credit_notes:
    primary key: CreditNoteID
    fields: Allocations(array), AppliedAmount(number), BrandingThemeID(string), CISDeduction(number), CISRate(number), Contact(object), CreditNoteID(string), CreditNoteNumber(string), CurrencyCode(string), CurrencyRate(number), Date(string), DueDate(string), FullyPaidOnDate(string), HasAttachments(boolean), HasErrors(boolean), InvoiceAddresses(array), LineAmountTypes(string), LineItems(array), Payments(array), Reference(string), RemainingCredit(number), SentToContact(boolean), Status(string), StatusAttributeString(string), SubTotal(number), Total(number), TotalTax(number), Type(string), UpdatedDateUTC(string), ValidationErrors(array), Warnings(array)
  currencies:
    primary key: Code
    fields: Code(string), Description(string)
  expense_claim:
    primary key: ExpenseClaimID
    fields: AmountDue(number), AmountPaid(number), ExpenseClaimID(string), PaymentDueDate(string), Payments(array), ReceiptID(string), Receipts(array), ReportingDate(string), Status(string), Total(number), UpdatedDateUTC(string), User(object), expense_claim_id(string)
  expense_claim_history:
    primary key: DateUTC
    fields: DateUTC(string)
  expense_claims:
    primary key: ExpenseClaimID
    fields: AmountDue(number), AmountPaid(number), ExpenseClaimID(string), PaymentDueDate(string), Payments(array), ReceiptID(string), Receipts(array), ReportingDate(string), Status(string), Total(number), UpdatedDateUTC(string), User(object)
  invoice:
    primary key: InvoiceID
    fields: AmountCredited(number), AmountDue(number), AmountPaid(number), Attachments(array), BrandingThemeID(string), CISDeduction(number), CISRate(number), Contact(object), CreditNotes(array), CurrencyCode(string), CurrencyRate(number), Date(string), DueDate(string), ExpectedPaymentDate(string), FullyPaidOnDate(string), HasAttachments(boolean), HasErrors(boolean), InvoiceAddresses(array), InvoiceID(string), InvoiceNumber(string), IsDiscounted(boolean), LineAmountTypes(string), LineItems(array), Overpayments(array), Payments(array), PlannedPaymentDate(string), Prepayments(array), Reference(string), RepeatingInvoiceID(string), SentToContact(boolean), Status(string), StatusAttributeString(string), SubTotal(number), Total(number), TotalDiscount(number), TotalTax(number), Type(string), UpdatedDateUTC(string), Url(string), ValidationErrors(array), Warnings(array), invoice_id(string)
  invoice_attachments:
    primary key: AttachmentID
    fields: AttachmentID(string), ContentLength(integer), FileName(string), IncludeOnline(boolean), MimeType(string), Url(string), invoice_id(string)
  invoice_history:
    primary key: DateUTC
    fields: DateUTC(string)
  invoice_reminders:
    primary key: Enabled
    fields: Enabled(boolean)
  item:
    primary key: ItemID
    fields: Code(string), Description(string), InventoryAssetAccountCode(string), IsPurchased(boolean), IsSold(boolean), IsTrackedAsInventory(boolean), ItemID(string), Name(string), PurchaseDescription(string), PurchaseDetails(object), QuantityOnHand(number), SalesDetails(object), StatusAttributeString(string), TotalCostPool(number), UpdatedDateUTC(string), ValidationErrors(array), item_id(string)
  item_history:
    primary key: DateUTC
    fields: DateUTC(string)
  journal:
    primary key: JournalID
    fields: CreatedDateUTC(string), JournalDate(string), JournalID(string), JournalLines(array), JournalNumber(integer), Reference(string), SourceID(string), SourceType(string), journal_id(string)
  journal_by_number:
    primary key: JournalID
    fields: CreatedDateUTC(string), JournalDate(string), JournalID(string), JournalLines(array), JournalNumber(integer), Reference(string), SourceID(string), SourceType(string), journal_number(string)
  journals:
    primary key: JournalID
    fields: CreatedDateUTC(string), JournalDate(string), JournalID(string), JournalLines(array), JournalNumber(integer), Reference(string), SourceID(string), SourceType(string)
  linked_transaction:
    primary key: ContactID
    fields: ContactID(string), LinkedTransactionID(string), SourceLineItemID(string), SourceTransactionID(string), SourceTransactionTypeCode(string), Status(string), TargetLineItemID(string), TargetTransactionID(string), Type(string), UpdatedDateUTC(string), ValidationErrors(array), linked_transaction_id(string)
  linked_transactions:
    primary key: ContactID
    fields: ContactID(string), LinkedTransactionID(string), SourceLineItemID(string), SourceTransactionID(string), SourceTransactionTypeCode(string), Status(string), TargetLineItemID(string), TargetTransactionID(string), Type(string), UpdatedDateUTC(string), ValidationErrors(array)
  manual_journal:
    primary key: ManualJournalID
    fields: Attachments(array), Date(string), HasAttachments(boolean), JournalLines(array), LineAmountTypes(string), ManualJournalID(string), Narration(string), ShowOnCashBasisReports(boolean), Status(string), StatusAttributeString(string), UpdatedDateUTC(string), Url(string), ValidationErrors(array), Warnings(array), manual_journal_id(string)
  manual_journal_attachments:
    primary key: AttachmentID
    fields: AttachmentID(string), ContentLength(integer), FileName(string), IncludeOnline(boolean), MimeType(string), Url(string), manual_journal_id(string)
  manual_journals:
    primary key: ManualJournalID
    fields: Attachments(array), Date(string), HasAttachments(boolean), JournalLines(array), LineAmountTypes(string), ManualJournalID(string), Narration(string), ShowOnCashBasisReports(boolean), Status(string), StatusAttributeString(string), UpdatedDateUTC(string), Url(string), ValidationErrors(array), Warnings(array)
  manual_journals_history:
    primary key: DateUTC
    fields: DateUTC(string)
  online_invoice:
    primary key: invoice_id
    fields: OnlineInvoiceUrl(string), invoice_id(string)
  organisation_actions:
    primary key: Name
    fields: Name(string), Status(string)
  organisation_cis_settings:
    primary key: organisation_id
    fields: CISContractorEnabled(boolean), CISSubContractorEnabled(boolean), Rate(number), organisation_id(string)
  organisations:
    primary key: OrganisationID
    fields: APIKey(string), Addresses(array), BaseCurrency(string), Class(string), CountryCode(string), CreatedDateUTC(string), DefaultPurchasesTax(string), DefaultSalesTax(string), Edition(string), EmployerIdentificationNumber(string), EndOfYearLockDate(string), ExternalLinks(array), FinancialYearEndDay(integer), FinancialYearEndMonth(integer), IsDemoCompany(boolean), LegalName(string), LineOfBusiness(string), Name(string), OrganisationEntityType(string), OrganisationID(string), OrganisationStatus(string), OrganisationType(string), PaymentTerms(object), PaysTax(boolean), PeriodLockDate(string), Phones(array), RegistrationNumber(string), SalesTaxBasis(string), SalesTaxPeriod(string), ShortCode(string), TaxNumber(string), Timezone(string), Version(string)
  overpayment:
    primary key: CurrencyCode
    fields: Allocations(array), AppliedAmount(number), Attachments(array), Contact(object), CurrencyCode(string), CurrencyRate(number), Date(string), HasAttachments(boolean), LineAmountTypes(string), LineItems(array), OverpaymentID(string), Payments(array), Reference(string), RemainingCredit(number), Status(string), SubTotal(number), Total(number), TotalTax(number), Type(string), UpdatedDateUTC(string), overpayment_id(string)
  overpayment_history:
    primary key: DateUTC
    fields: DateUTC(string)
  overpayments:
    primary key: CurrencyCode
    fields: Allocations(array), AppliedAmount(number), Attachments(array), Contact(object), CurrencyCode(string), CurrencyRate(number), Date(string), HasAttachments(boolean), LineAmountTypes(string), LineItems(array), OverpaymentID(string), Payments(array), Reference(string), RemainingCredit(number), Status(string), SubTotal(number), Total(number), TotalTax(number), Type(string), UpdatedDateUTC(string)
  payment:
    primary key: PaymentID
    fields: Account(object), Amount(number), BankAccountNumber(string), BankAmount(number), BatchPayment(object), BatchPaymentID(string), Code(string), CreditNote(object), CreditNoteNumber(string), CurrencyRate(number), Date(string), Details(string), HasAccount(boolean), HasValidationErrors(boolean), Invoice(object), InvoiceNumber(string), IsReconciled(boolean), Overpayment(object), Particulars(string), PaymentID(string), PaymentType(string), Prepayment(object), Reference(string), Status(string), StatusAttributeString(string), UpdatedDateUTC(string), ValidationErrors(array), Warnings(array), payment_id(string)
  payment_history:
    primary key: DateUTC
    fields: DateUTC(string)
  payment_services:
    primary key: PaymentServiceID
    fields: PayNowText(string), PaymentServiceID(string), PaymentServiceName(string), PaymentServiceType(string), PaymentServiceUrl(string), ValidationErrors(array)
  prepayment:
    primary key: BrandingThemeID
    fields: Allocations(array), AppliedAmount(number), Attachments(array), BrandingThemeID(string), Contact(object), CurrencyCode(string), CurrencyRate(number), Date(string), HasAttachments(boolean), InvoiceNumber(string), LineAmountTypes(string), LineItems(array), Payments(array), PrepaymentID(string), Reference(string), RemainingCredit(number), Status(string), SubTotal(number), Total(number), TotalTax(number), Type(string), UpdatedDateUTC(string), prepayment_id(string)
  prepayment_history:
    primary key: DateUTC
    fields: DateUTC(string)
  prepayments:
    primary key: BrandingThemeID
    fields: Allocations(array), AppliedAmount(number), Attachments(array), BrandingThemeID(string), Contact(object), CurrencyCode(string), CurrencyRate(number), Date(string), HasAttachments(boolean), InvoiceNumber(string), LineAmountTypes(string), LineItems(array), Payments(array), PrepaymentID(string), Reference(string), RemainingCredit(number), Status(string), SubTotal(number), Total(number), TotalTax(number), Type(string), UpdatedDateUTC(string)
  purchase_order:
    primary key: PurchaseOrderID
    fields: Attachments(array), AttentionTo(string), BrandingThemeID(string), Contact(object), CurrencyCode(string), CurrencyRate(number), Date(string), DeliveryAddress(string), DeliveryDate(string), DeliveryInstructions(string), ExpectedArrivalDate(string), HasAttachments(boolean), LineAmountTypes(string), LineItems(array), PurchaseOrderID(string), PurchaseOrderNumber(string), Reference(string), SentToContact(boolean), Status(string), StatusAttributeString(string), SubTotal(number), Telephone(string), Total(number), TotalDiscount(number), TotalTax(number), UpdatedDateUTC(string), ValidationErrors(array), Warnings(array), purchase_order_id(string)
  purchase_order_attachments:
    primary key: AttachmentID
    fields: AttachmentID(string), ContentLength(integer), FileName(string), IncludeOnline(boolean), MimeType(string), Url(string), purchase_order_id(string)
  purchase_order_by_number:
    primary key: PurchaseOrderID
    fields: Attachments(array), AttentionTo(string), BrandingThemeID(string), Contact(object), CurrencyCode(string), CurrencyRate(number), Date(string), DeliveryAddress(string), DeliveryDate(string), DeliveryInstructions(string), ExpectedArrivalDate(string), HasAttachments(boolean), LineAmountTypes(string), LineItems(array), PurchaseOrderID(string), PurchaseOrderNumber(string), Reference(string), SentToContact(boolean), Status(string), StatusAttributeString(string), SubTotal(number), Telephone(string), Total(number), TotalDiscount(number), TotalTax(number), UpdatedDateUTC(string), ValidationErrors(array), Warnings(array), purchase_order_number(string)
  purchase_order_history:
    primary key: DateUTC
    fields: DateUTC(string)
  purchase_orders:
    primary key: PurchaseOrderID
    fields: Attachments(array), AttentionTo(string), BrandingThemeID(string), Contact(object), CurrencyCode(string), CurrencyRate(number), Date(string), DeliveryAddress(string), DeliveryDate(string), DeliveryInstructions(string), ExpectedArrivalDate(string), HasAttachments(boolean), LineAmountTypes(string), LineItems(array), PurchaseOrderID(string), PurchaseOrderNumber(string), Reference(string), SentToContact(boolean), Status(string), StatusAttributeString(string), SubTotal(number), Telephone(string), Total(number), TotalDiscount(number), TotalTax(number), UpdatedDateUTC(string), ValidationErrors(array), Warnings(array)
  quote:
    primary key: QuoteID
    fields: BrandingThemeID(string), Contact(object), CurrencyCode(string), CurrencyRate(number), Date(string), DateString(string), ExpiryDate(string), ExpiryDateString(string), LineAmountTypes(string), LineItems(array), QuoteID(string), QuoteNumber(string), Reference(string), Status(string), StatusAttributeString(string), SubTotal(number), Summary(string), Terms(string), Title(string), Total(number), TotalDiscount(number), TotalTax(number), UpdatedDateUTC(string), ValidationErrors(array), quote_id(string)
  quote_attachments:
    primary key: AttachmentID
    fields: AttachmentID(string), ContentLength(integer), FileName(string), IncludeOnline(boolean), MimeType(string), Url(string), quote_id(string)
  quote_history:
    primary key: DateUTC
    fields: DateUTC(string)
  quotes:
    primary key: QuoteID
    fields: BrandingThemeID(string), Contact(object), CurrencyCode(string), CurrencyRate(number), Date(string), DateString(string), ExpiryDate(string), ExpiryDateString(string), LineAmountTypes(string), LineItems(array), QuoteID(string), QuoteNumber(string), Reference(string), Status(string), StatusAttributeString(string), SubTotal(number), Summary(string), Terms(string), Title(string), Total(number), TotalDiscount(number), TotalTax(number), UpdatedDateUTC(string), ValidationErrors(array)
  receipt:
    primary key: ReceiptID
    fields: Attachments(array), Contact(object), Date(string), HasAttachments(boolean), LineAmountTypes(string), LineItems(array), ReceiptID(string), ReceiptNumber(string), Reference(string), Status(string), SubTotal(number), Total(number), TotalTax(number), UpdatedDateUTC(string), Url(string), User(object), ValidationErrors(array), Warnings(array), receipt_id(string)
  receipt_attachments:
    primary key: AttachmentID
    fields: AttachmentID(string), ContentLength(integer), FileName(string), IncludeOnline(boolean), MimeType(string), Url(string), receipt_id(string)
  receipt_history:
    primary key: DateUTC
    fields: DateUTC(string)
  receipts:
    primary key: ReceiptID
    fields: Attachments(array), Contact(object), Date(string), HasAttachments(boolean), LineAmountTypes(string), LineItems(array), ReceiptID(string), ReceiptNumber(string), Reference(string), Status(string), SubTotal(number), Total(number), TotalTax(number), UpdatedDateUTC(string), Url(string), User(object), ValidationErrors(array), Warnings(array)
  repeating_invoice:
    primary key: ID
    fields: ApprovedForSending(boolean), Attachments(array), BrandingThemeID(string), Contact(object), CurrencyCode(string), HasAttachments(boolean), ID(string), IncludePDF(boolean), LineAmountTypes(string), LineItems(array), MarkAsSent(boolean), Reference(string), RepeatingInvoiceID(string), Schedule(object), SendCopy(boolean), Status(string), SubTotal(number), Total(number), TotalTax(number), Type(string), repeating_invoice_id(string)
  repeating_invoice_attachments:
    primary key: AttachmentID
    fields: AttachmentID(string), ContentLength(integer), FileName(string), IncludeOnline(boolean), MimeType(string), Url(string), repeating_invoice_id(string)
  repeating_invoice_history:
    primary key: DateUTC
    fields: DateUTC(string)
  repeating_invoices:
    primary key: ID
    fields: ApprovedForSending(boolean), Attachments(array), BrandingThemeID(string), Contact(object), CurrencyCode(string), HasAttachments(boolean), ID(string), IncludePDF(boolean), LineAmountTypes(string), LineItems(array), MarkAsSent(boolean), Reference(string), RepeatingInvoiceID(string), Schedule(object), SendCopy(boolean), Status(string), SubTotal(number), Total(number), TotalTax(number), Type(string)
  report_aged_payables_by_contact:
    primary key: ReportID
    fields: Fields(array), ReportDate(string), ReportID(string), ReportName(string), ReportTitle(string), ReportTitles(array), ReportType(string), Rows(array), UpdatedDateUTC(string)
  report_aged_receivables_by_contact:
    primary key: ReportID
    fields: Fields(array), ReportDate(string), ReportID(string), ReportName(string), ReportTitle(string), ReportTitles(array), ReportType(string), Rows(array), UpdatedDateUTC(string)
  report_balance_sheet:
    primary key: ReportID
    fields: Fields(array), ReportDate(string), ReportID(string), ReportName(string), ReportTitle(string), ReportTitles(array), ReportType(string), Rows(array), UpdatedDateUTC(string)
  report_bank_summary:
    primary key: ReportID
    fields: Fields(array), ReportDate(string), ReportID(string), ReportName(string), ReportTitle(string), ReportTitles(array), ReportType(string), Rows(array), UpdatedDateUTC(string)
  report_budget_summary:
    primary key: ReportID
    fields: Fields(array), ReportDate(string), ReportID(string), ReportName(string), ReportTitle(string), ReportTitles(array), ReportType(string), Rows(array), UpdatedDateUTC(string)
  report_executive_summary:
    primary key: ReportID
    fields: Fields(array), ReportDate(string), ReportID(string), ReportName(string), ReportTitle(string), ReportTitles(array), ReportType(string), Rows(array), UpdatedDateUTC(string)
  report_from_id:
    primary key: ReportID
    fields: Fields(array), ReportDate(string), ReportID(string), ReportName(string), ReportTitle(string), ReportTitles(array), ReportType(string), Rows(array), UpdatedDateUTC(string), report_id(string)
  report_profit_and_loss:
    primary key: ReportID
    fields: Fields(array), ReportDate(string), ReportID(string), ReportName(string), ReportTitle(string), ReportTitles(array), ReportType(string), Rows(array), UpdatedDateUTC(string)
  report_ten_ninety_nine:
    primary key: ReportName
    fields: Contacts(array), ReportDate(string), ReportName(string), ReportTitle(string), ReportType(string), UpdatedDateUTC(string)
  report_trial_balance:
    primary key: ReportID
    fields: Fields(array), ReportDate(string), ReportID(string), ReportName(string), ReportTitle(string), ReportTitles(array), ReportType(string), Rows(array), UpdatedDateUTC(string)
  reports_list:
    primary key: ReportID
    fields: Fields(array), ReportDate(string), ReportID(string), ReportName(string), ReportTitle(string), ReportTitles(array), ReportType(string), Rows(array), UpdatedDateUTC(string)
  tax_rate_by_tax_type:
    primary key: TaxType
    fields: CanApplyToAssets(boolean), CanApplyToEquity(boolean), CanApplyToExpenses(boolean), CanApplyToLiabilities(boolean), CanApplyToRevenue(boolean), DisplayTaxRate(number), EffectiveRate(number), Name(string), ReportTaxType(string), Status(string), TaxComponents(array), TaxType(string), tax_type(string)
  tax_rates:
    primary key: TaxType
    fields: CanApplyToAssets(boolean), CanApplyToEquity(boolean), CanApplyToExpenses(boolean), CanApplyToLiabilities(boolean), CanApplyToRevenue(boolean), DisplayTaxRate(number), EffectiveRate(number), Name(string), ReportTaxType(string), Status(string), TaxComponents(array), TaxType(string)
  tracking_categories:
    primary key: TrackingCategoryID
    fields: Name(string), Option(string), Options(array), Status(string), TrackingCategoryID(string), TrackingOptionID(string)
  tracking_category:
    primary key: TrackingCategoryID
    fields: Name(string), Option(string), Options(array), Status(string), TrackingCategoryID(string), TrackingOptionID(string), tracking_category_id(string)
  user:
    primary key: UserID
    fields: EmailAddress(string), FirstName(string), IsSubscriber(boolean), LastName(string), OrganisationRole(string), UpdatedDateUTC(string), UserID(string), user_id(string)
  users:
    primary key: UserID
    fields: EmailAddress(string), FirstName(string), IsSubscriber(boolean), LastName(string), OrganisationRole(string), UpdatedDateUTC(string), UserID(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_account:
    endpoint: PUT Accounts
    risk: creates Xero account resources in the connected tenant; approval required
  delete_account:
    endpoint: DELETE Accounts/{{ record.account_id }}
    required fields: account_id
    risk: deletes Xero account resources in the connected tenant; approval required
  update_account:
    endpoint: POST Accounts/{{ record.account_id }}
    required fields: account_id
    risk: mutates Xero account resources in the connected tenant; approval required
  upsert_bank_transactions:
    endpoint: POST BankTransactions
    risk: mutates Xero bank transactions resources in the connected tenant; approval required
  create_bank_transactions:
    endpoint: PUT BankTransactions
    risk: creates Xero bank transactions resources in the connected tenant; approval required
  update_bank_transaction:
    endpoint: POST BankTransactions/{{ record.bank_transaction_id }}
    required fields: bank_transaction_id
    risk: mutates Xero bank transaction resources in the connected tenant; approval required
  create_bank_transaction_history_record:
    endpoint: PUT BankTransactions/{{ record.bank_transaction_id }}/History
    required fields: bank_transaction_id, HistoryRecords
    risk: creates Xero bank transaction history record resources in the connected tenant; approval required
  create_bank_transfer:
    endpoint: PUT BankTransfers
    risk: creates Xero bank transfer resources in the connected tenant; approval required
  delete_bank_transfers:
    endpoint: POST BankTransfers
    required fields: BankTransfers
    risk: Destructive Xero Accounting API action: sets one or more bank transfers to DELETED; reverse ETL must preview records and require explicit approval before execute.
  delete_bank_transfer:
    endpoint: POST BankTransfers/{{ record.bank_transfer_id }}
    required fields: bank_transfer_id, Status
    risk: Destructive Xero Accounting API action: sets a bank transfer to DELETED; reverse ETL must preview the resolved bank_transfer_id and require explicit approval before execute.
  create_bank_transfer_history_record:
    endpoint: PUT BankTransfers/{{ record.bank_transfer_id }}/History
    required fields: bank_transfer_id, HistoryRecords
    risk: creates Xero bank transfer history record resources in the connected tenant; approval required
  delete_batch_payment:
    endpoint: POST BatchPayments
    required fields: Status, BatchPaymentID
    risk: deletes Xero batch payment resources in the connected tenant; approval required
  create_batch_payment:
    endpoint: PUT BatchPayments
    risk: creates Xero batch payment resources in the connected tenant; approval required
  delete_batch_payment_by_url_param:
    endpoint: POST BatchPayments/{{ record.batch_payment_id }}
    required fields: batch_payment_id, Status
    risk: deletes Xero batch payment by url param resources in the connected tenant; approval required
  create_batch_payment_history_record:
    endpoint: PUT BatchPayments/{{ record.batch_payment_id }}/History
    required fields: batch_payment_id, HistoryRecords
    risk: creates Xero batch payment history record resources in the connected tenant; approval required
  create_branding_theme_payment_services:
    endpoint: POST BrandingThemes/{{ record.branding_theme_id }}/PaymentServices
    required fields: branding_theme_id
    risk: creates Xero branding theme payment services resources in the connected tenant; approval required
  create_contact_group:
    endpoint: PUT ContactGroups
    risk: creates Xero contact group resources in the connected tenant; approval required
  update_contact_group:
    endpoint: POST ContactGroups/{{ record.contact_group_id }}
    required fields: contact_group_id
    risk: mutates Xero contact group resources in the connected tenant; approval required
  delete_contact_group_contacts:
    endpoint: DELETE ContactGroups/{{ record.contact_group_id }}/Contacts
    required fields: contact_group_id
    risk: deletes Xero contact group contacts resources in the connected tenant; approval required
  create_contact_group_contacts:
    endpoint: PUT ContactGroups/{{ record.contact_group_id }}/Contacts
    required fields: contact_group_id
    risk: creates Xero contact group contacts resources in the connected tenant; approval required
  delete_contact_group_contact:
    endpoint: DELETE ContactGroups/{{ record.contact_group_id }}/Contacts/{{ record.contact_id }}
    required fields: contact_group_id, contact_id
    risk: deletes Xero contact group contact resources in the connected tenant; approval required
  upsert_contacts:
    endpoint: POST Contacts
    risk: mutates Xero contacts resources in the connected tenant; approval required
  create_contacts:
    endpoint: PUT Contacts
    risk: creates Xero contacts resources in the connected tenant; approval required
  update_contact:
    endpoint: POST Contacts/{{ record.contact_id }}
    required fields: contact_id
    risk: mutates Xero contact resources in the connected tenant; approval required
  create_contact_history:
    endpoint: PUT Contacts/{{ record.contact_id }}/History
    required fields: contact_id, HistoryRecords
    risk: creates Xero contact history resources in the connected tenant; approval required
  upsert_credit_notes:
    endpoint: POST CreditNotes
    risk: mutates Xero credit notes resources in the connected tenant; approval required
  create_credit_notes:
    endpoint: PUT CreditNotes
    risk: creates Xero credit notes resources in the connected tenant; approval required
  update_credit_note:
    endpoint: POST CreditNotes/{{ record.credit_note_id }}
    required fields: credit_note_id
    risk: mutates Xero credit note resources in the connected tenant; approval required
  create_credit_note_allocation:
    endpoint: PUT CreditNotes/{{ record.credit_note_id }}/Allocations
    required fields: credit_note_id
    risk: creates Xero credit note allocation resources in the connected tenant; approval required
  delete_credit_note_allocations:
    endpoint: DELETE CreditNotes/{{ record.credit_note_id }}/Allocations/{{ record.allocation_id }}
    required fields: credit_note_id, allocation_id
    risk: deletes Xero credit note allocations resources in the connected tenant; approval required
  create_credit_note_history:
    endpoint: PUT CreditNotes/{{ record.credit_note_id }}/History
    required fields: credit_note_id, HistoryRecords
    risk: creates Xero credit note history resources in the connected tenant; approval required
  create_currency:
    endpoint: PUT Currencies
    risk: creates Xero currency resources in the connected tenant; approval required
  create_expense_claims:
    endpoint: PUT ExpenseClaims
    risk: creates Xero expense claims resources in the connected tenant; approval required
  update_expense_claim:
    endpoint: POST ExpenseClaims/{{ record.expense_claim_id }}
    required fields: expense_claim_id
    risk: mutates Xero expense claim resources in the connected tenant; approval required
  create_expense_claim_history:
    endpoint: PUT ExpenseClaims/{{ record.expense_claim_id }}/History
    required fields: expense_claim_id, HistoryRecords
    risk: creates Xero expense claim history resources in the connected tenant; approval required
  upsert_invoices:
    endpoint: POST Invoices
    risk: mutates Xero invoices resources in the connected tenant; approval required
  create_invoices:
    endpoint: PUT Invoices
    risk: creates Xero invoices resources in the connected tenant; approval required
  update_invoice:
    endpoint: POST Invoices/{{ record.invoice_id }}
    required fields: invoice_id
    risk: mutates Xero invoice resources in the connected tenant; approval required
  email_invoice:
    endpoint: POST Invoices/{{ record.invoice_id }}/Email
    required fields: invoice_id
    risk: executes Xero email invoice resources in the connected tenant; approval required
  create_invoice_history:
    endpoint: PUT Invoices/{{ record.invoice_id }}/History
    required fields: invoice_id, HistoryRecords
    risk: creates Xero invoice history resources in the connected tenant; approval required
  upsert_items:
    endpoint: POST Items
    risk: mutates Xero items resources in the connected tenant; approval required
  create_items:
    endpoint: PUT Items
    risk: creates Xero items resources in the connected tenant; approval required
  delete_item:
    endpoint: DELETE Items/{{ record.item_id }}
    required fields: item_id
    risk: deletes Xero item resources in the connected tenant; approval required
  update_item:
    endpoint: POST Items/{{ record.item_id }}
    required fields: item_id
    risk: mutates Xero item resources in the connected tenant; approval required
  create_item_history:
    endpoint: PUT Items/{{ record.item_id }}/History
    required fields: item_id, HistoryRecords
    risk: creates Xero item history resources in the connected tenant; approval required
  create_linked_transaction:
    endpoint: PUT LinkedTransactions
    risk: creates Xero linked transaction resources in the connected tenant; approval required
  delete_linked_transaction:
    endpoint: DELETE LinkedTransactions/{{ record.linked_transaction_id }}
    required fields: linked_transaction_id
    risk: deletes Xero linked transaction resources in the connected tenant; approval required
  update_linked_transaction:
    endpoint: POST LinkedTransactions/{{ record.linked_transaction_id }}
    required fields: linked_transaction_id
    risk: mutates Xero linked transaction resources in the connected tenant; approval required
  upsert_manual_journals:
    endpoint: POST ManualJournals
    risk: mutates Xero manual journals resources in the connected tenant; approval required
  create_manual_journals:
    endpoint: PUT ManualJournals
    risk: creates Xero manual journals resources in the connected tenant; approval required
  update_manual_journal:
    endpoint: POST ManualJournals/{{ record.manual_journal_id }}
    required fields: manual_journal_id
    risk: mutates Xero manual journal resources in the connected tenant; approval required
  create_manual_journal_history_record:
    endpoint: PUT ManualJournals/{{ record.manual_journal_id }}/History
    required fields: manual_journal_id, HistoryRecords
    risk: creates Xero manual journal history record resources in the connected tenant; approval required
  create_overpayment_allocations:
    endpoint: PUT Overpayments/{{ record.overpayment_id }}/Allocations
    required fields: overpayment_id
    risk: creates Xero overpayment allocations resources in the connected tenant; approval required
  delete_overpayment_allocations:
    endpoint: DELETE Overpayments/{{ record.overpayment_id }}/Allocations/{{ record.allocation_id }}
    required fields: overpayment_id, allocation_id
    risk: deletes Xero overpayment allocations resources in the connected tenant; approval required
  create_overpayment_history:
    endpoint: PUT Overpayments/{{ record.overpayment_id }}/History
    required fields: overpayment_id, HistoryRecords
    risk: creates Xero overpayment history resources in the connected tenant; approval required
  create_payment:
    endpoint: POST Payments
    risk: creates Xero payment resources in the connected tenant; approval required
  create_payments:
    endpoint: PUT Payments
    risk: creates Xero payments resources in the connected tenant; approval required
  delete_payment:
    endpoint: POST Payments/{{ record.payment_id }}
    required fields: payment_id, Status
    risk: deletes Xero payment resources in the connected tenant; approval required
  create_payment_history:
    endpoint: PUT Payments/{{ record.payment_id }}/History
    required fields: payment_id, HistoryRecords
    risk: creates Xero payment history resources in the connected tenant; approval required
  create_payment_service:
    endpoint: PUT PaymentServices
    risk: creates Xero payment service resources in the connected tenant; approval required
  create_prepayment_allocations:
    endpoint: PUT Prepayments/{{ record.prepayment_id }}/Allocations
    required fields: prepayment_id
    risk: creates Xero prepayment allocations resources in the connected tenant; approval required
  delete_prepayment_allocations:
    endpoint: DELETE Prepayments/{{ record.prepayment_id }}/Allocations/{{ record.allocation_id }}
    required fields: prepayment_id, allocation_id
    risk: deletes Xero prepayment allocations resources in the connected tenant; approval required
  create_prepayment_history:
    endpoint: PUT Prepayments/{{ record.prepayment_id }}/History
    required fields: prepayment_id, HistoryRecords
    risk: creates Xero prepayment history resources in the connected tenant; approval required
  upsert_purchase_orders:
    endpoint: POST PurchaseOrders
    risk: mutates Xero purchase orders resources in the connected tenant; approval required
  create_purchase_orders:
    endpoint: PUT PurchaseOrders
    risk: creates Xero purchase orders resources in the connected tenant; approval required
  update_purchase_order:
    endpoint: POST PurchaseOrders/{{ record.purchase_order_id }}
    required fields: purchase_order_id
    risk: mutates Xero purchase order resources in the connected tenant; approval required
  create_purchase_order_history:
    endpoint: PUT PurchaseOrders/{{ record.purchase_order_id }}/History
    required fields: purchase_order_id, HistoryRecords
    risk: creates Xero purchase order history resources in the connected tenant; approval required
  upsert_quotes:
    endpoint: POST Quotes
    risk: mutates Xero quotes resources in the connected tenant; approval required
  create_quotes:
    endpoint: PUT Quotes
    risk: creates Xero quotes resources in the connected tenant; approval required
  update_quote:
    endpoint: POST Quotes/{{ record.quote_id }}
    required fields: quote_id
    risk: mutates Xero quote resources in the connected tenant; approval required
  create_quote_history:
    endpoint: PUT Quotes/{{ record.quote_id }}/History
    required fields: quote_id, HistoryRecords
    risk: creates Xero quote history resources in the connected tenant; approval required
  create_receipt:
    endpoint: PUT Receipts
    risk: creates Xero receipt resources in the connected tenant; approval required
  update_receipt:
    endpoint: POST Receipts/{{ record.receipt_id }}
    required fields: receipt_id
    risk: mutates Xero receipt resources in the connected tenant; approval required
  create_receipt_history:
    endpoint: PUT Receipts/{{ record.receipt_id }}/History
    required fields: receipt_id, HistoryRecords
    risk: creates Xero receipt history resources in the connected tenant; approval required
  upsert_repeating_invoices:
    endpoint: POST RepeatingInvoices
    risk: mutates Xero repeating invoices resources in the connected tenant; approval required
  create_repeating_invoices:
    endpoint: PUT RepeatingInvoices
    risk: creates Xero repeating invoices resources in the connected tenant; approval required
  update_repeating_invoice:
    endpoint: POST RepeatingInvoices/{{ record.repeating_invoice_id }}
    required fields: repeating_invoice_id
    risk: mutates Xero repeating invoice resources in the connected tenant; approval required
  create_repeating_invoice_history:
    endpoint: PUT RepeatingInvoices/{{ record.repeating_invoice_id }}/History
    required fields: repeating_invoice_id, HistoryRecords
    risk: creates Xero repeating invoice history resources in the connected tenant; approval required
  setup_organisation:
    endpoint: POST Setup
    risk: mutates Xero setup resources in the connected tenant; approval required
  update_tax_rate:
    endpoint: POST TaxRates
    risk: mutates Xero tax rate resources in the connected tenant; approval required
  create_tax_rates:
    endpoint: PUT TaxRates
    risk: creates Xero tax rates resources in the connected tenant; approval required
  create_tracking_category:
    endpoint: PUT TrackingCategories
    risk: creates Xero tracking category resources in the connected tenant; approval required
  delete_tracking_category:
    endpoint: DELETE TrackingCategories/{{ record.tracking_category_id }}
    required fields: tracking_category_id
    risk: deletes Xero tracking category resources in the connected tenant; approval required
  update_tracking_category:
    endpoint: POST TrackingCategories/{{ record.tracking_category_id }}
    required fields: tracking_category_id
    risk: mutates Xero tracking category resources in the connected tenant; approval required
  create_tracking_options:
    endpoint: PUT TrackingCategories/{{ record.tracking_category_id }}/Options
    required fields: tracking_category_id
    risk: creates Xero tracking options resources in the connected tenant; approval required
  delete_tracking_options:
    endpoint: DELETE TrackingCategories/{{ record.tracking_category_id }}/Options/{{ record.tracking_option_id }}
    required fields: tracking_category_id, tracking_option_id
    risk: deletes Xero tracking options resources in the connected tenant; approval required
  update_tracking_options:
    endpoint: POST TrackingCategories/{{ record.tracking_category_id }}/Options/{{ record.tracking_option_id }}
    required fields: tracking_category_id, tracking_option_id
    risk: mutates Xero tracking options resources in the connected tenant; approval required

SECURITY
  read risk: external Xero Accounting API read of financial, contact, report, attachment-metadata, history, and organisation data
  write risk: creates, updates, emails, sets up, and deletes Xero Accounting API resources in the connected tenant
  approval: reverse ETL writes require plan preview and approval token; delete actions are marked destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run bounded Xero Accounting API streams, report reads, and approved reverse-ETL writes.
  Usage: pm connectors command xero <group> <action> --credential <name> [flags] --json
  Source CLI: Xero Accounting API (https://raw.githubusercontent.com/XeroAPI/Xero-OpenAPI/master/xero_accounting.yaml)
  Global flags:
    --credential (string): Named Xero credential; secrets are loaded from the credential store and never from prompt text.
    --json (boolean): Emit machine-readable JSON output.
    --max-bytes (integer): Clamp direct-read response size; report reads are capped by the operation definition.
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Bounded report direct reads
  ETL stream shortcuts
  Other Commands
    streams invoices - Read the Xero `invoices` stream through the ETL engine. [intent=etl availability=implemented stream=invoices]
    streams contacts - Read the Xero `contacts` stream through the ETL engine. [intent=etl availability=implemented stream=contacts]
    streams accounts - Read the Xero `accounts` stream through the ETL engine. [intent=etl availability=implemented stream=accounts]
    streams bank_transactions - Read the Xero `bank_transactions` stream through the ETL engine. [intent=etl availability=implemented stream=bank_transactions]
    streams items - Read the Xero `items` stream through the ETL engine. [intent=etl availability=implemented stream=items]
    streams payments - Read the Xero `payments` stream through the ETL engine. [intent=etl availability=implemented stream=payments]
    streams account - Read the Xero `account` stream through the ETL engine. [intent=etl availability=implemented stream=account]
    streams account_attachments - Read the Xero `account_attachments` stream through the ETL engine. [intent=etl availability=implemented stream=account_attachments]
    streams bank_transaction - Read the Xero `bank_transaction` stream through the ETL engine. [intent=etl availability=implemented stream=bank_transaction]
    streams bank_transaction_attachments - Read the Xero `bank_transaction_attachments` stream through the ETL engine. [intent=etl availability=implemented stream=bank_transaction_attachments]
    streams bank_transactions_history - Read the Xero `bank_transactions_history` stream through the ETL engine. [intent=etl availability=implemented stream=bank_transactions_history]
    streams bank_transfer - Read the Xero `bank_transfer` stream through the ETL engine. [intent=etl availability=implemented stream=bank_transfer]
    reports ten_ninety_nine - Retrieve reports for 1099 [intent=direct_read availability=implemented operation=xero.get_report_ten_ninety_nine]; approval: none; risk: bounded read; report responses are capped at 16 MiB and redacted before JSON output; flags: --page, --page-cursor
    reports aged_payables_by_contact - Retrieves report for aged payables by contact [intent=direct_read availability=implemented operation=xero.get_report_aged_payables_by_contact]; approval: none; risk: bounded read; report responses are capped at 16 MiB and redacted before JSON output; flags: --page, --page-cursor
    reports aged_receivables_by_contact - Retrieves report for aged receivables by contact [intent=direct_read availability=implemented operation=xero.get_report_aged_receivables_by_contact]; approval: none; risk: bounded read; report responses are capped at 16 MiB and redacted before JSON output; flags: --page, --page-cursor
    reports balance_sheet - Retrieves report for balancesheet [intent=direct_read availability=implemented operation=xero.get_report_balance_sheet]; approval: none; risk: bounded read; report responses are capped at 16 MiB and redacted before JSON output; flags: --page, --page-cursor
    reports bank_summary - Retrieves report for bank summary [intent=direct_read availability=implemented operation=xero.get_report_bank_summary]; approval: none; risk: bounded read; report responses are capped at 16 MiB and redacted before JSON output; flags: --page, --page-cursor
    reports get - Retrieves a specific report using a unique ReportID [intent=direct_read availability=implemented operation=xero.get_report_from_id]; approval: none; risk: bounded read; report responses are capped at 16 MiB and redacted before JSON output; flags: --page, --page-cursor
    reports budget_summary - Retrieves report for budget summary [intent=direct_read availability=implemented operation=xero.get_report_budget_summary]; approval: none; risk: bounded read; report responses are capped at 16 MiB and redacted before JSON output; flags: --page, --page-cursor
    reports executive_summary - Retrieves report for executive summary [intent=direct_read availability=implemented operation=xero.get_report_executive_summary]; approval: none; risk: bounded read; report responses are capped at 16 MiB and redacted before JSON output; flags: --page, --page-cursor
    reports list - Retrieves a list of the organistaions unique reports that require a uuid to fetch [intent=direct_read availability=implemented operation=xero.get_reports_list]; approval: none; risk: bounded read; report responses are capped at 16 MiB and redacted before JSON output; flags: --page, --page-cursor
    reports profit_and_loss - Retrieves report for profit and loss [intent=direct_read availability=implemented operation=xero.get_report_profit_and_loss]; approval: none; risk: bounded read; report responses are capped at 16 MiB and redacted before JSON output; flags: --page, --page-cursor
    reports trial_balance - Retrieves report for trial balance [intent=direct_read availability=implemented operation=xero.get_report_trial_balance]; approval: none; risk: bounded read; report responses are capped at 16 MiB and redacted before JSON output; flags: --page, --page-cursor
    attachments get_account_attachment_by_id - Retrieves a specific attachment from a specific account using a unique attachment Id [intent=direct_read availability=planned operation=xero.get_account_attachment_by_id]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_account_attachment_by_file_name - Retrieves an attachment for a specific account by filename [intent=direct_read availability=planned operation=xero.get_account_attachment_by_file_name]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments update_account_attachment_by_file_name - Updates attachment on a specific account by filename [intent=reverse_etl availability=planned operation=xero.update_account_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments create_account_attachment_by_file_name - Creates an attachment on a specific account [intent=reverse_etl availability=planned operation=xero.create_account_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments get_bank_transaction_attachment_by_id - Retrieves specific attachments from a specific BankTransaction using a unique attachment Id [intent=direct_read availability=planned operation=xero.get_bank_transaction_attachment_by_id]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_bank_transaction_attachment_by_file_name - Retrieves a specific attachment from a specific bank transaction by filename [intent=direct_read availability=planned operation=xero.get_bank_transaction_attachment_by_file_name]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments update_bank_transaction_attachment_by_file_name - Updates a specific attachment from a specific bank transaction by filename [intent=reverse_etl availability=planned operation=xero.update_bank_transaction_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments create_bank_transaction_attachment_by_file_name - Creates an attachment for a specific bank transaction by filename [intent=reverse_etl availability=planned operation=xero.create_bank_transaction_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments get_bank_transfer_attachment_by_id - Retrieves a specific attachment from a specific bank transfer using a unique attachment ID [intent=direct_read availability=planned operation=xero.get_bank_transfer_attachment_by_id]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_bank_transfer_attachment_by_file_name - Retrieves a specific attachment on a specific bank transfer by file name [intent=direct_read availability=planned operation=xero.get_bank_transfer_attachment_by_file_name]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments update_bank_transfer_attachment_by_file_name - updateBankTransferAttachmentByFileName [intent=reverse_etl availability=planned operation=xero.update_bank_transfer_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments create_bank_transfer_attachment_by_file_name - createBankTransferAttachmentByFileName [intent=reverse_etl availability=planned operation=xero.create_bank_transfer_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments get_contact_attachment_by_id - Retrieves a specific attachment from a specific contact using a unique attachment Id [intent=direct_read availability=planned operation=xero.get_contact_attachment_by_id]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_contact_attachment_by_file_name - Retrieves a specific attachment from a specific contact by file name [intent=direct_read availability=planned operation=xero.get_contact_attachment_by_file_name]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments update_contact_attachment_by_file_name - updateContactAttachmentByFileName [intent=reverse_etl availability=planned operation=xero.update_contact_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments create_contact_attachment_by_file_name - createContactAttachmentByFileName [intent=reverse_etl availability=planned operation=xero.create_contact_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments get_credit_note_attachment_by_id - Retrieves a specific attachment from a specific credit note using a unique attachment Id [intent=direct_read availability=planned operation=xero.get_credit_note_attachment_by_id]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_credit_note_attachment_by_file_name - Retrieves a specific attachment on a specific credit note by file name [intent=direct_read availability=planned operation=xero.get_credit_note_attachment_by_file_name]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments update_credit_note_attachment_by_file_name - Updates attachments on a specific credit note by file name [intent=reverse_etl availability=planned operation=xero.update_credit_note_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments create_credit_note_attachment_by_file_name - Creates an attachment for a specific credit note [intent=reverse_etl availability=planned operation=xero.create_credit_note_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments get_credit_note_as_pdf - Retrieves credit notes as PDF files [intent=direct_read availability=planned operation=xero.get_credit_note_as_pdf]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_invoice_as_pdf - Retrieves invoices or purchase bills as PDF files [intent=direct_read availability=planned operation=xero.get_invoice_as_pdf]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_invoice_attachment_by_id - Retrieves a specific attachment from a specific invoices or purchase bills by using a unique attachment Id [intent=direct_read availability=planned operation=xero.get_invoice_attachment_by_id]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_invoice_attachment_by_file_name - Retrieves an attachment from a specific invoice or purchase bill by filename [intent=direct_read availability=planned operation=xero.get_invoice_attachment_by_file_name]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments update_invoice_attachment_by_file_name - Updates an attachment from a specific invoices or purchase bill by filename [intent=reverse_etl availability=planned operation=xero.update_invoice_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments create_invoice_attachment_by_file_name - Creates an attachment for a specific invoice or purchase bill by filename [intent=reverse_etl availability=planned operation=xero.create_invoice_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments get_manual_journal_attachment_by_id - Allows you to retrieve a specific attachment from a specific manual journal using a unique attachment Id [intent=direct_read availability=planned operation=xero.get_manual_journal_attachment_by_id]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_manual_journal_attachment_by_file_name - Retrieves a specific attachment from a specific manual journal by file name [intent=direct_read availability=planned operation=xero.get_manual_journal_attachment_by_file_name]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments update_manual_journal_attachment_by_file_name - Updates a specific attachment from a specific manual journal by file name [intent=reverse_etl availability=planned operation=xero.update_manual_journal_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments create_manual_journal_attachment_by_file_name - Creates a specific attachment for a specific manual journal by file name [intent=reverse_etl availability=planned operation=xero.create_manual_journal_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments get_purchase_order_as_pdf - Retrieves specific purchase order as PDF files using a unique purchase order Id [intent=direct_read availability=planned operation=xero.get_purchase_order_as_pdf]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_purchase_order_attachment_by_id - Retrieves specific attachment for a specific purchase order using a unique attachment Id [intent=direct_read availability=planned operation=xero.get_purchase_order_attachment_by_id]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_purchase_order_attachment_by_file_name - Retrieves a specific attachment for a specific purchase order by filename [intent=direct_read availability=planned operation=xero.get_purchase_order_attachment_by_file_name]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments update_purchase_order_attachment_by_file_name - Updates a specific attachment for a specific purchase order by filename [intent=reverse_etl availability=planned operation=xero.update_purchase_order_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments create_purchase_order_attachment_by_file_name - Creates attachment for a specific purchase order [intent=reverse_etl availability=planned operation=xero.create_purchase_order_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments get_quote_as_pdf - Retrieves a specific quote as a PDF file using a unique quote Id [intent=direct_read availability=planned operation=xero.get_quote_as_pdf]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_quote_attachment_by_id - Retrieves a specific attachment from a specific quote using a unique attachment Id [intent=direct_read availability=planned operation=xero.get_quote_attachment_by_id]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_quote_attachment_by_file_name - Retrieves a specific attachment from a specific quote by filename [intent=direct_read availability=planned operation=xero.get_quote_attachment_by_file_name]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments update_quote_attachment_by_file_name - Updates a specific attachment from a specific quote by filename [intent=reverse_etl availability=planned operation=xero.update_quote_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments create_quote_attachment_by_file_name - Creates attachment for a specific quote [intent=reverse_etl availability=planned operation=xero.create_quote_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments get_receipt_attachment_by_id - Retrieves a specific attachments from a specific expense claim receipts by using a unique attachment Id [intent=direct_read availability=planned operation=xero.get_receipt_attachment_by_id]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_receipt_attachment_by_file_name - Retrieves a specific attachment from a specific expense claim receipts by file name [intent=direct_read availability=planned operation=xero.get_receipt_attachment_by_file_name]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments update_receipt_attachment_by_file_name - Updates a specific attachment on a specific expense claim receipts by file name [intent=reverse_etl availability=planned operation=xero.update_receipt_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments create_receipt_attachment_by_file_name - Creates an attachment on a specific expense claim receipts by file name [intent=reverse_etl availability=planned operation=xero.create_receipt_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments get_repeating_invoice_attachment_by_id - Retrieves a specific attachment from a specific repeating invoice [intent=direct_read availability=planned operation=xero.get_repeating_invoice_attachment_by_id]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments get_repeating_invoice_attachment_by_file_name - Retrieves a specific attachment from a specific repeating invoices by file name [intent=direct_read availability=planned operation=xero.get_repeating_invoice_attachment_by_file_name]; approval: none; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.; flags: --page, --page-cursor
    attachments update_repeating_invoice_attachment_by_file_name - Updates a specific attachment from a specific repeating invoices by file name [intent=reverse_etl availability=planned operation=xero.update_repeating_invoice_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
    attachments create_repeating_invoice_attachment_by_file_name - Creates an attachment from a specific repeating invoices by file name [intent=reverse_etl availability=planned operation=xero.create_repeating_invoice_attachment_by_file_name]; approval: binary/file uploads require plan, preview, explicit approval, and payload digest approval before execution; risk: blocked shared-runtime binary/file transfer dependency; notes: Operation is bounded in operations.json; no raw path/body/method escape hatch is exposed.
  Help topics:
    xero reports - Bounded direct reads for Xero Accounting report endpoints with typed query/path flags.
    xero reverse-etl - Reverse ETL writes use typed schemas and require plan, preview, explicit approval, and execute.
    xero attachments - Attachment metadata streams are available; binary/PDF download and upload execution is blocked on the shared binary/file runner.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect xero

  # Inspect as structured JSON
  pm connectors inspect xero --json

AGENT WORKFLOW
  - Run pm connectors inspect xero before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
