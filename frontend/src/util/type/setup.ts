// * use generic response type from generated backend
import type { ResponseSuccessResponse } from '$/util/backend/backend'

export type Setup = {
	profile: {
		id?: string
		name?: string
		email?: string
		userId?: string
	}
	initialized: boolean
	reload: () => Promise<void>
}
