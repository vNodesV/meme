use crate::msg::{MintingFeesResponse, OriginalAdminResponse};
use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};

// this is a mapping of address to token_id
pub const PRIMARY_ALIASES: Map<&Addr, String> = Map::new("aliases");

// this is fees info
pub const MINTING_FEES_INFO: Item<MintingFeesResponse> = Item::new("minting_fees");

// this is original admin info
pub const ORIGINAL_ADMIN_INFO: Item<OriginalAdminResponse> = Item::new("original_admin");