#!/usr/bin/env node

// This checkpoint is the rendered-browser supplement for the seven Square
// groups whose fetched HTML has no static endpoint cards. The rows were
// extracted from the provider-rendered DOM with chrome-devtools-axi and retain
// the provider's endpoint-page URL as the operation citation.
import { readFileSync, writeFileSync } from 'node:fs';

const progressPath = new URL('../crawl-progress.json', import.meta.url);
const root = 'https://developer.squareup.com/reference/square/';
const renderedGroups = {
  'checkout-api': [
    ['GET', '/v2/online-checkout/location-settings/{location_id}', 'retrieve-location-settings'], ['PUT', '/v2/online-checkout/location-settings/{location_id}', 'update-location-settings'], ['GET', '/v2/online-checkout/merchant-settings', 'retrieve-merchant-settings'], ['PUT', '/v2/online-checkout/merchant-settings', 'update-merchant-settings'], ['GET', '/v2/online-checkout/payment-links', 'list-payment-links'], ['POST', '/v2/online-checkout/payment-links', 'create-payment-link'], ['DELETE', '/v2/online-checkout/payment-links/{id}', 'delete-payment-link'], ['GET', '/v2/online-checkout/payment-links/{id}', 'retrieve-payment-link'], ['PUT', '/v2/online-checkout/payment-links/{id}', 'update-payment-link'],
  ],
  'gift-cards-api': [
    ['GET', '/v2/gift-cards', 'list-gift-cards'], ['POST', '/v2/gift-cards', 'create-gift-card'], ['POST', '/v2/gift-cards/from-gan', 'retrieve-gift-card-from-gan'], ['POST', '/v2/gift-cards/from-nonce', 'retrieve-gift-card-from-nonce'], ['POST', '/v2/gift-cards/{gift_card_id}/link-customer', 'link-customer-to-gift-card'], ['POST', '/v2/gift-cards/{gift_card_id}/unlink-customer', 'unlink-customer-from-gift-card'], ['GET', '/v2/gift-cards/{id}', 'retrieve-gift-card'],
  ],
  'inventory-api': [
    ['GET', '/v2/inventory/adjustment-reasons', 'list-inventory-adjustment-reasons'], ['POST', '/v2/inventory/adjustment-reasons/create', 'create-inventory-adjustment-reason'], ['POST', '/v2/inventory/adjustment-reasons/delete', 'delete-inventory-adjustment-reason'], ['POST', '/v2/inventory/adjustment-reasons/restore', 'restore-inventory-adjustment-reason'], ['POST', '/v2/inventory/adjustment-reasons/retrieve', 'retrieve-inventory-adjustment-reason'], ['PUT', '/v2/inventory/adjustment-reasons/update', 'update-inventory-adjustment-reason'], ['PUT', '/v2/inventory/adjustments/update', 'update-inventory-adjustment'], ['GET', '/v2/inventory/adjustments/{adjustment_id}', 'retrieve-inventory-adjustment'], ['POST', '/v2/inventory/changes/batch-create', 'batch-change-inventory'], ['POST', '/v2/inventory/changes/batch-retrieve', 'batch-retrieve-inventory-changes'], ['POST', '/v2/inventory/counts/batch-retrieve', 'batch-retrieve-inventory-counts'], ['GET', '/v2/inventory/physical-counts/{physical_count_id}', 'retrieve-inventory-physical-count'], ['GET', '/v2/inventory/{catalog_object_id}', 'retrieve-inventory-count'],
  ],
  'locations-api': [
    ['GET', '/v2/locations', 'list-locations'], ['POST', '/v2/locations', 'create-location'], ['GET', '/v2/locations/{location_id}', 'retrieve-location'], ['PUT', '/v2/locations/{location_id}', 'update-location'],
  ],
  'merchant-custom-attributes-api': [
    ['GET', '/v2/merchants/custom-attribute-definitions', 'list-merchant-custom-attribute-definitions'], ['POST', '/v2/merchants/custom-attribute-definitions', 'create-merchant-custom-attribute-definition'], ['DELETE', '/v2/merchants/custom-attribute-definitions/{key}', 'delete-merchant-custom-attribute-definition'], ['GET', '/v2/merchants/custom-attribute-definitions/{key}', 'retrieve-merchant-custom-attribute-definition'], ['PUT', '/v2/merchants/custom-attribute-definitions/{key}', 'update-merchant-custom-attribute-definition'], ['POST', '/v2/merchants/custom-attributes/bulk-delete', 'bulk-delete-merchant-custom-attributes'], ['POST', '/v2/merchants/custom-attributes/bulk-upsert', 'bulk-upsert-merchant-custom-attributes'], ['GET', '/v2/merchants/{merchant_id}/custom-attributes', 'list-merchant-custom-attributes'], ['DELETE', '/v2/merchants/{merchant_id}/custom-attributes/{key}', 'delete-merchant-custom-attribute'], ['GET', '/v2/merchants/{merchant_id}/custom-attributes/{key}', 'retrieve-merchant-custom-attribute'], ['POST', '/v2/merchants/{merchant_id}/custom-attributes/{key}', 'upsert-merchant-custom-attribute'],
  ],
  'payouts-api': [
    ['GET', '/v2/payouts', 'list-payouts'], ['GET', '/v2/payouts/{payout_id}', 'get-payout'], ['GET', '/v2/payouts/{payout_id}/payout-entries', 'list-payout-entries'],
  ],
  'vendors-api': [
    ['POST', '/v2/vendors/bulk-create', 'bulk-create-vendors'], ['POST', '/v2/vendors/bulk-retrieve', 'bulk-retrieve-vendors'], ['PUT', '/v2/vendors/bulk-update', 'bulk-update-vendors'], ['POST', '/v2/vendors/create', 'create-vendor'], ['POST', '/v2/vendors/search', 'search-vendors'], ['GET', '/v2/vendors/{vendor_id}', 'retrieve-vendor'], ['PUT', '/v2/vendors/{vendor_id}', 'update-vendor'],
  ],
};

const progress = JSON.parse(readFileSync(progressPath, 'utf8'));
const crawl = progress.crawls.square;
for (const [group, rows] of Object.entries(renderedGroups)) {
  const url = new URL(group, root).href;
  const page = crawl.group_pages?.[url];
  if (!page) throw new Error(`missing persisted group page ${url}`);
  if (page.operations?.length) throw new Error(`refusing to replace already extracted operations for ${url}`);
  page.operations = rows.map(([method, path, operation]) => ({ method, path, source_url: new URL(`${group}/${operation}`, root).href }));
  page.extraction = 'browser_rendered_dom';
}
crawl.groups_extracted = Object.values(crawl.group_pages).filter(page => page.operations?.length).length;
crawl.state = 'in_progress';
crawl.coverage_confidence = 'partial';
crawl.operations_found = null;
crawl.last_error = null;
crawl.resume_strategy = 'Browser-rendered operation cards have been persisted for the groups that lack static cards. Rerun square-reference-crawl.mjs to re-deduplicate only after every group has operations.';
progress.updated_at = new Date().toISOString();
writeFileSync(progressPath, `${JSON.stringify(progress, null, 2)}\n`);
console.log(`square browser extraction checkpoint: ${crawl.groups_extracted}/${crawl.groups_total} groups with operations`);
