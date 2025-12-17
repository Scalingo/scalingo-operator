package helpers

import (
	"context"

	"github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/scalingo-operator/internal/domain"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type SecretManager struct {
	client             client.Client
	databaseMetaObject metav1.Object
}

func NewSecretManager(client client.Client, databaseMetaObject metav1.Object) *SecretManager {
	return &SecretManager{
		client:             client,
		databaseMetaObject: databaseMetaObject,
	}
}

func (m SecretManager) GetSecret(ctx context.Context, secret domain.Secret) (string, error) {
	if secret.Namespace == "" {
		return "", errors.New(ctx, "empty namespace")
	}
	if secret.Name == "" {
		return "", errors.New(ctx, "empty name")
	}
	if secret.Key == "" {
		return "", errors.New(ctx, "empty key")
	}

	coreSecret := &corev1.Secret{}
	err := m.client.Get(ctx, client.ObjectKey{Namespace: secret.Namespace, Name: secret.Name}, coreSecret)
	if err != nil {
		return "", errors.Wrap(ctx, err, "get auth secret")
	}
	data, ok := coreSecret.Data[secret.Key]
	if !ok {
		return "", errors.New(ctx, "get auth secret key")
	}
	return string(data), nil
}

func (m SecretManager) SetSecret(ctx context.Context, secret domain.Secret) error {
	if secret.Namespace == "" {
		return errors.New(ctx, "empty namespace")
	}
	if secret.Name == "" {
		return errors.New(ctx, "empty name")
	}
	if secret.Key == "" {
		return errors.New(ctx, "empty key")
	}
	if secret.Value == "" {
		return errors.New(ctx, "empty value")
	}

	coreSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: secret.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, m.client, coreSecret, func() error {
		if coreSecret.Data == nil {
			coreSecret.Data = make(map[string][]byte)
		}

		coreSecret.Data[secret.Key] = []byte(secret.Value)

		err := controllerutil.SetControllerReference(m.databaseMetaObject, coreSecret, m.client.Scheme())
		if err != nil {
			return errors.Wrap(ctx, err, "set controller reference on secret")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(ctx, err, "create or update secret")
	}
	return nil
}

func ComposeConnectionURLName(prefix, defaultName string) string {
	if prefix == "" {
		return defaultName
	}
	return prefix + connectionURLNameSuffix
}

const connectionURLNameSuffix = "_URL"
