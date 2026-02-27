mod contract_tests;
mod error;
pub mod execute;
pub mod msg;
pub mod query;
pub mod state;
pub mod utils;

use cosmwasm_std::{ensure_eq, to_json_binary, Empty};

use cw2::set_contract_version;
use execute::{
    execute_instantiate, mint, set_admin_address,
    update_minting_fees, CONTRACT_NAME, CONTRACT_VERSION, update_origin_image, update_origin_admin
};
use query::{
    address_of, contract_info, get_parent_id, get_parent_nft_info, original_admin
};

pub use crate::msg::{ExecuteMsg, Extension, InstantiateMsg, MigrateMsg, QueryMsg};

pub use crate::error::ContractError;

pub type Cw721MetadataContract = cw721_base::Cw721Contract<Extension, Empty, Empty, Empty>;

pub mod entry {

    use super::*;

    use cosmwasm_std::entry_point;
    use cosmwasm_std::{Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};

    #[cfg_attr(not(feature = "library"), entry_point)]
    pub fn instantiate(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        msg: InstantiateMsg,
    ) -> StdResult<Response> {
        let tract = Cw721MetadataContract::default();
        execute_instantiate(tract, deps, env, info, msg)
    }

    #[cfg_attr(not(feature = "library"), entry_point)]
    pub fn execute(
        deps: DepsMut,
        env: Env,
        info: MessageInfo,
        msg: ExecuteMsg,
    ) -> Result<Response, ContractError> {
        let tract = Cw721MetadataContract::default();
        match msg {
            ExecuteMsg::UpdateMintingFees(msg) => update_minting_fees(tract, deps, env, info, msg),
            ExecuteMsg::UpdateOriginImage(msg) => update_origin_image(tract, deps, env, info, msg),
            ExecuteMsg::UpdateOriginAdmin(msg) => update_origin_admin(tract, deps, env, info, msg),
            ExecuteMsg::Mint(msg) => mint(tract, deps, env, info, msg),
            ExecuteMsg::SetAdminAddress { admin_address } => {
                set_admin_address(tract, deps, env, info, admin_address)
            }
            _ => tract
                .execute(deps, env, info, msg.into())
                .map_err(ContractError::Base),
        }
    }

    #[cfg_attr(not(feature = "library"), entry_point)]
    pub fn query(deps: Deps, env: Env, msg: QueryMsg) -> StdResult<Binary> {
        let tract = Cw721MetadataContract::default();

        match msg {
            QueryMsg::OriginAdminInfo {} => to_json_binary(&original_admin(deps)?),
            QueryMsg::ContractInfo {} => to_json_binary(&contract_info(deps)?),
            QueryMsg::AddressOf { token_id } => to_json_binary(&address_of(tract, deps, token_id)?),
            QueryMsg::GetParentId { token_id } => to_json_binary(&get_parent_id(tract, deps, token_id)?),
            QueryMsg::GetParentInfo { token_id } => {
                to_json_binary(&get_parent_nft_info(tract, deps, token_id)?)
            }
            _ => tract.query(deps, env, msg.into()).map_err(|err| err),
        }
    }

    #[cfg_attr(not(feature = "library"), entry_point)]
    pub fn migrate(deps: DepsMut, _env: Env, msg: MigrateMsg) -> Result<Response, ContractError> {
        ensure_eq!(
            msg.target_version,
            CONTRACT_VERSION,
            ContractError::Unauthorized {}
        );

        set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;
        Ok(Response::new().add_attribute("action", "migrate"))
    }
}
