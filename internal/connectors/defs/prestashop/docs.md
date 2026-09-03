# PrestaShop Connector

## Overview

Reads PrestaShop customers, orders, products, addresses, and carts through the PrestaShop Webservice REST API.

Readable streams: `customers`, `orders`, `products`, `addresses`, `carts`.

Service API documentation: https://devdocs.prestashop-project.org/9/webservice/.

## Auth setup

Connection fields:

- `access_key` (required, secret, string); Your PrestaShop access key. See <a href="https://devdocs.prestashop.com/1.7/webservice/tutorials/creating-access/#create-an-access-key"> the docs </a> for info on how to obtain this.
- `start_date` (optional, string); The Start date in the format YYYY-MM-DD.
- `url` (required, string); Required PrestaShop shop HTTPS origin.

Authentication uses declared mode(s): `basic`.

## Execution contract

Connection check: `GET /customers`
Check query: `display`=`full`; `limit`=`1`; `output_format`=`JSON`.

## Streams notes

- `customers`: `GET /customers`; records `customers.customer`
  - Query: `display`=`full`; `filter[date_upd]`=`[{{ incremental.lower_bound }},]`; `output_format`=`JSON`; `sort`=`[date_upd_ASC]`.
  - Pagination: `offset_count`.
  - Incremental cursor: `date_upd`.
- `orders`: `GET /orders`; records `orders.order`
  - Query: `display`=`full`; `filter[date_upd]`=`[{{ incremental.lower_bound }},]`; `output_format`=`JSON`; `sort`=`[date_upd_ASC]`.
  - Pagination: `offset_count`.
  - Incremental cursor: `date_upd`.
- `products`: `GET /products`; records `products.product`
  - Query: `display`=`full`; `filter[date_upd]`=`[{{ incremental.lower_bound }},]`; `output_format`=`JSON`; `sort`=`[date_upd_ASC]`.
  - Pagination: `offset_count`.
  - Incremental cursor: `date_upd`.
- `addresses`: `GET /addresses`; records `addresses.address`
  - Query: `display`=`full`; `filter[date_upd]`=`[{{ incremental.lower_bound }},]`; `output_format`=`JSON`; `sort`=`[date_upd_ASC]`.
  - Pagination: `offset_count`.
  - Incremental cursor: `date_upd`.
- `carts`: `GET /carts`; records `carts.cart`
  - Query: `display`=`full`; `filter[date_upd]`=`[{{ incremental.lower_bound }},]`; `output_format`=`JSON`; `sort`=`[date_upd_ASC]`.
  - Pagination: `offset_count`.
  - Incremental cursor: `date_upd`.

## Write actions & risks

This connector's write surface is declared separately in the rendered execution bundle.
