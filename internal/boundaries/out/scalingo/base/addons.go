package scalingo

import (
	"context"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
	errors "github.com/Scalingo/go-utils/errors/v3"
)

// getAddonIDFromDatabase resolves the addon ID from a database name by calling the API.
//
// Taken from `cli` project:
// https://github.com/Scalingo/cli/blob/093b3f099210a9f1a5a1bab49980b94576c299b9/detect/app.go#L23
func (c *client) getAddonIDFromDatabase(ctx context.Context, databaseName string) (string, error) {
	// AddonsList works for both apps and DBNG databases (same API endpoint).
	// A DBNG database is modeled as an app with a single addon (itself),
	// whereas an application can have multiple addons (postgresql, redis, etc.).
	// If multiple addons are returned, the ID is likely an application, not a database.
	addons, err := c.scClient.AddonsList(ctx, databaseName)
	if err != nil {
		return "", errors.Wrap(ctx, err, "list addons")
	}

	if len(addons) == 0 {
		return "", errors.Newf(ctx, "no addon found for database %s", databaseName)
	}

	if len(addons) > 1 {
		return "", errors.Newf(ctx, "multiple addons found for %s, it may be an application", databaseName)
	}

	return addons[0].ID, nil
}

// findPlanID retrieves a plan ID from its name by calling the API.
//
// Taken from `cli` project:
// https://github.com/Scalingo/cli/blob/master/utils/plans.go#L11
func (c *client) findPlanID(ctx context.Context, addonID, planName string) (string, error) {
	plans, err := c.scClient.AddonProviderPlansList(ctx, addonID, scalingoapi.AddonProviderPlansListOpts{})
	if err != nil {
		return "", errors.Wrapf(ctx, err, "list addon %s plans", addonID)
	}
	for _, p := range plans {
		if p.Name == planName {
			return p.ID, nil
		}
	}
	return "", errors.Newf(ctx, "no plan %s found for addon %s", planName, addonID)
}
