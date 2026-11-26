import { apiGet, apiPut } from './http'
export type ProviderConnection = {
 providerCode: string; adapterType: string; name: string; apiBaseUrl: string; environment: string
 enabled: boolean; supportedCountries: string; supportedCurrencies: string; callbackIPAllowlist: string[]
 version: number; source: string; credentialsConfigured: boolean; credentialFields: Record<string, boolean>
 marketAllowed: boolean; callbackPath: string
}
export type ConnectionList = {
 list: ProviderConnection[]; market: { countryCode: string; currencyCode: string }; callbackBaseUrl: string
 adapters: { type: string; credentialFields: string[]; requiredFields: string[] }[]
 countries: { countryCode: string; countryName: string; currencyCode: string }[]
}
export type ConnectionInput = Omit<ProviderConnection,'source'|'credentialsConfigured'|'credentialFields'|'marketAllowed'|'callbackPath'> & {
 credentials: Record<string,string>; clearCredentials: boolean; reason: string
}
export const fetchProviderConnections = () => apiGet<ConnectionList>('/admin/game-provider-connections')
export const saveProviderConnection = (input: ConnectionInput) => apiPut<ProviderConnection>(`/admin/game-provider-connections/${encodeURIComponent(input.providerCode)}`,input)
