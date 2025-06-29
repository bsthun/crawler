export type Setup = {
	profile: {
		userId?: string
		name?: string
		email?: string
		photoUrl?: string
		isAdmin?: boolean
	}
	initialized: boolean
	reload: () => Promise<void>
}
