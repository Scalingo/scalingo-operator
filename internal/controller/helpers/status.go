package helpers

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Databases status annotation.
const DatabaseAnnotationIsRunning = "databases.scalingo.com/db-is-running"

// Helper functions to read and modify operator status through its
// Meta data: annotations and status conditions.

func IsDatabaseInitialized(conditions []metav1.Condition) bool {
	return meta.IsStatusConditionFalse(conditions, string(DatabaseStatusConditionAvailable))
}

func IsDatabaseProvisioning(conditions []metav1.Condition) bool {
	return meta.IsStatusConditionTrue(conditions, string(DatabaseStatusConditionProvisioning))
}

func IsDatabaseAvailable(conditions []metav1.Condition) bool {
	return meta.IsStatusConditionTrue(conditions, string(DatabaseStatusConditionAvailable))
}

func IsDatabaseRunning(dbMeta metav1.ObjectMeta) bool {
	return metav1.HasAnnotation(dbMeta, DatabaseAnnotationIsRunning) &&
		dbMeta.Annotations[DatabaseAnnotationIsRunning] == annotationValueTrue
}

func IsDatabaseDeletionRequested(dbMeta metav1.ObjectMeta) bool {
	return !dbMeta.DeletionTimestamp.IsZero()
}

func SetDatabaseInitialStatus(conditions *[]metav1.Condition) {
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:    string(DatabaseStatusConditionAvailable),
		Status:  metav1.ConditionFalse,
		Reason:  reasonNotAvailable,
		Message: msgNotAvailable,
	})

	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:    string(DatabaseStatusConditionProvisioning),
		Status:  metav1.ConditionFalse,
		Reason:  reasonNotprovisioned,
		Message: msgNotProvisioned,
	})
}

func SetDatabaseStatusProvisioning(conditions *[]metav1.Condition) {
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:    string(DatabaseStatusConditionProvisioning),
		Status:  metav1.ConditionTrue,
		Reason:  reasonProvisioning,
		Message: msgProvisioning,
	})
}

func SetDatabaseStatusProvisioned(conditions *[]metav1.Condition) {
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:    string(DatabaseStatusConditionAvailable),
		Status:  metav1.ConditionTrue,
		Reason:  reasonAvailable,
		Message: msgAvailable,
	})
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:    string(DatabaseStatusConditionProvisioning),
		Status:  metav1.ConditionFalse,
		Reason:  reasonProvisioned,
		Message: msgProvisioned,
	})
}

func SetDatabaseIsNotRunning(dbMeta *metav1.ObjectMeta) {
	metav1.SetMetaDataAnnotation(dbMeta, DatabaseAnnotationIsRunning, annotationValueFalse)
}

func SetDatabaseIsRunning(dbMeta *metav1.ObjectMeta) {
	metav1.SetMetaDataAnnotation(dbMeta, DatabaseAnnotationIsRunning, annotationValueTrue)
}

// Private constants.
const (
	reasonNotAvailable   = "DatabaseNotAvailable"
	reasonAvailable      = "DatabaseAvailable"
	reasonNotprovisioned = "DatabaseNotprovisioned"
	reasonProvisioning   = "DatabaseProvisioning"
	reasonProvisioned    = "Databaseprovisioned"

	msgNotAvailable   = "The database is not yet available on Scalingo."
	msgAvailable      = "The database is available on Scalingo."
	msgNotProvisioned = "The database is not yet provisioned on Scalingo."
	msgProvisioning   = "The database is being provisioned on Scalingo."
	msgProvisioned    = "The database is provisioned on Scalingo."

	annotationValueTrue  = "true"
	annotationValueFalse = "false"
)
