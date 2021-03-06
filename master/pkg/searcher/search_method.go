package searcher

import (
	"encoding/json"

	"github.com/determined-ai/determined/master/pkg/model"
	"github.com/determined-ai/determined/master/pkg/nprand"
	"github.com/determined-ai/determined/master/pkg/workload"
)

type context struct {
	rand    *nprand.State
	hparams model.Hyperparameters
}

// SearchMethod is the interface for hyper-parameter tuning methods. Implementations of this
// interface should use pointer receivers to ensure interface equality is calculated through pointer
// equality.
type SearchMethod interface {
	// initialOperations returns a set of initial operations that the searcher would like to take.
	// This should be called only once after the searcher has been created.
	initialOperations(ctx context) ([]Operation, error)
	// trialCreated informs the searcher that a trial has been created as a result of a Create
	// operation.
	trialCreated(ctx context, requestID model.RequestID) ([]Operation, error)
	// trainCompleted informs the searcher that the training workload initiated by the same searcher
	// has completed. It returns any new operations as a result of this workload completing.
	trainCompleted(ctx context, requestID model.RequestID, train Train) ([]Operation, error)
	// checkpointCompleted informs the searcher that the checkpoint workload initiated by the same
	// searcher has completed. It returns any new operations as a result of this workload
	// completing.
	checkpointCompleted(
		ctx context, requestID model.RequestID, checkpoint Checkpoint, metrics workload.CheckpointMetrics,
	) ([]Operation, error)
	// validationCompleted informs the searcher that the validation workload initiated by the same
	// searcher has completed. It returns any new operations as a result of this workload
	// completing.
	validationCompleted(
		ctx context, requestID model.RequestID, validate Validate, metrics workload.ValidationMetrics,
	) ([]Operation, error)
	// trialClosed informs the searcher that the trial has been closed as a result of a Close
	// operation.
	trialClosed(ctx context, requestID model.RequestID) ([]Operation, error)
	// progress returns experiment progress as a float between 0.0 and 1.0. As search methods
	// receive completed workloads, they should internally track progress.
	progress(totalUnitsCompleted float64) float64
	// trialExitedEarly informs the searcher that the trial has exited earlier than expected.
	trialExitedEarly(
		ctx context, requestID model.RequestID, exitedReason workload.ExitedReason,
	) ([]Operation, error)
	// TODO: refactor as model.Snapshotter interface or something
	model.Snapshotter
	model.InUnits
}

// NewSearchMethod returns a new search method for the provided searcher configuration.
func NewSearchMethod(c model.SearcherConfig) SearchMethod {
	switch {
	case c.SingleConfig != nil:
		return newSingleSearch(*c.SingleConfig)
	case c.RandomConfig != nil:
		return newRandomSearch(*c.RandomConfig)
	case c.GridConfig != nil:
		return newGridSearch(*c.GridConfig)
	case c.SyncHalvingConfig != nil:
		return newSyncHalvingSearch(*c.SyncHalvingConfig)
	case c.AdaptiveConfig != nil:
		return newAdaptiveSearch(*c.AdaptiveConfig)
	case c.AdaptiveSimpleConfig != nil:
		return newAdaptiveSimpleSearch(*c.AdaptiveSimpleConfig)
	case c.AsyncHalvingConfig != nil:
		return newAsyncHalvingSearch(*c.AsyncHalvingConfig)
	case c.AdaptiveASHAConfig != nil:
		return newAdaptiveASHASearch(*c.AdaptiveASHAConfig)
	case c.PBTConfig != nil:
		return newPBTSearch(*c.PBTConfig)
	default:
		panic("no searcher type specified")
	}
}

type defaultSearchMethod struct{}

func (defaultSearchMethod) trialCreated(context, model.RequestID) ([]Operation, error) {
	return nil, nil
}
func (defaultSearchMethod) trainCompleted(context, model.RequestID, Train) ([]Operation, error) {
	return nil, nil
}

func (defaultSearchMethod) checkpointCompleted(
	context, model.RequestID, Checkpoint, workload.CheckpointMetrics,
) ([]Operation, error) {
	return nil, nil
}
func (defaultSearchMethod) validationCompleted(
	context, model.RequestID, Validate, workload.ValidationMetrics,
) ([]Operation, error) {
	return nil, nil
}
func (defaultSearchMethod) trialClosed(context, model.RequestID) ([]Operation, error) {
	return nil, nil
}

func (defaultSearchMethod) trialExitedEarly( //nolint: unused
	context, model.RequestID, workload.ExitedReason) ([]Operation, error) {
	return []Operation{Shutdown{Failure: true}}, nil
}

// save is the default implementation, used by stateless searchers like random and grid.
func (defaultSearchMethod) Snapshot() (json.RawMessage, error) {
	return nil, nil
}

// load is the default implementation, used by stateless searchers like random and grid.
func (defaultSearchMethod) Restore(json.RawMessage) error {
	return nil
}
