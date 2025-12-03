package policy

func (policy Policy) Update(newPolicy Policy) (Policy, error) {
	if policy.Id == "" {
		return Policy{}, ErrPolicyIDRequired
	}

	newPolicyObj, err := newPolicy.Create()
	if err != nil {
		return Policy{}, err
	}

	err = policy.Delete()
	if err != nil {
		// TODO: WHat if this delete also errors out?
		newPolicyObj.Delete()
		return Policy{}, err
	}

	return newPolicyObj, nil
}
