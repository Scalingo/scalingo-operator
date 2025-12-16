package scalingo

import (
	"context"

	errors "github.com/Scalingo/go-utils/errors/v3"
)

func (c *client) FindApplicationVariable(ctx context.Context, appID, varName string) (string, error) {
	variables, err := c.scClient.VariablesListWithoutAlias(ctx, appID)
	if err != nil {
		return "", errors.Wrap(ctx, err, "find application variable")
	}

	variable, found := variables.Contains(varName)
	if !found {
		return "", errors.Newf(ctx, "variable %s not found", varName)
	}

	return variable.Value, nil
}
