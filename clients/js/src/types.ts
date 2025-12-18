/**
 * TypeScript types for the Message Delivery Service API.
 * re-exported from auto-generated openapi-typescript definition.
 */
import type { components } from "./schema.js";

// Helper alias to access schemas
type Schemas = components["schemas"];

export type EmailContact = Schemas["EmailContact"];
export type EmailRequest = Schemas["EmailRequest"];
export type SuccessResponse = Schemas["SuccessResponse"];
export type ErrorResponse = Schemas["ErrorResponse"];
export type SmsRecipient = Schemas["SmsRecipient"];
export type SmsRequest = Schemas["SmsRequest"];
export type SmsSuccessResponse = Schemas["SmsSuccessResponse"];
