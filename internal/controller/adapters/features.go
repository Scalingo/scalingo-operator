package adapters

import (
	"github.com/Scalingo/scalingo-operator/api/v1alpha1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func toFeatures(networkSpec *v1alpha1.NetworkingSpec) domain.DatabaseFeatures {
	if networkSpec == nil || !networkSpec.InternetAccess.Enabled {
		return nil
	}
	return domain.DatabaseFeatures{
		domain.DatabaseFeaturePubliclyAvailable: domain.DatabaseFeatureStatusActivated,
	}
}
