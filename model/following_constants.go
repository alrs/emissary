package model

// FollowMethodActivityPub represents the ActivityPub subscription
// https://www.w3.org/TR/activitypub/
const FollowMethodActivityPub = "ACTIVITYPUB"

// FollowMethodPoll represents a subscription that must be polled for updates
const FollowMethodPoll = "POLL"

// FollowMethodWebSub represents a WebSub subscription
// https://websub.rocks
const FollowMethodWebSub = "WEBSUB"

// FollowingStatusNew represents a new following that has not yet been polled
const FollowingStatusNew = "NEW"

// FollowingStatusLoading represents a following that is being loaded for the first time
const FollowingStatusLoading = "LOADING"

// FollowingStatusSuccess represents a following that has successfully loaded
const FollowingStatusSuccess = "SUCCESS"

// FollowingStatusFailure represents a following that has failed to load
const FollowingStatusFailure = "FAILURE"
