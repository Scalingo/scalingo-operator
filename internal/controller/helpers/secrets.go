package helpers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Scalingo/go-utils/errors/v2"
)

func GetSecret(ctx context.Context, reader client.Reader, namespace, name, key string) (string, error) {
	authSecret := &corev1.Secret{}
	err := reader.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, authSecret)
	if err != nil {
		return "", errors.Wrap(ctx, err, "get auth secret")
	}
	data, ok := authSecret.Data[key]
	if !ok {
		return "", errors.New(ctx, "get auth secret key")
	}
	return string(data), nil
}
