#!/usr/bin/env fish

function _log_info
    echo (set_color cyan)"→"(set_color normal) $argv
end

function _log_success
    echo (set_color green)"✓"(set_color normal) $argv
end

function _log_warn
    echo (set_color yellow)"!"(set_color normal) $argv
end

function _log_error
    echo (set_color red)"✗"(set_color normal) $argv
end

function _prompt
    printf (set_color brblack)"["(set_color magenta)"?"(set_color brblack)"] "(set_color normal)"%s " $argv[1]
    read response
    echo $response
end

function createRelease -d "Create a new release"
    # Check for svu
    if not command -q svu
        _log_error "svu is not installed. Install with: go install github.com/caarlos0/svu@latest"
        return 1
    end

    # Check for uncommitted changes
    if test -n "$(git status --porcelain)"
        _log_warn "You have uncommitted changes:"
        git status --short
        echo
        set -l commit_response (_prompt "Commit changes before release? [y/N]")
        if string match -qi "y*" $commit_response
            set -l commit_msg (_prompt "Commit message:")
            if test -z "$commit_msg"
                _log_error "Commit message cannot be empty"
                return 1
            end
            git add -A
            git commit -m "$commit_msg"
            _log_success "Changes committed"
        else
            _log_error "Cannot create release with uncommitted changes"
            return 1
        end
    end

    # Get next version
    set -l nextVersion (svu next)
    if test -z "$nextVersion"
        _log_error "Could not determine next version"
        return 1
    end

    set -l currentVersion (svu current 2>/dev/null; or echo "none")
    echo
    _log_info "Current version:" (set_color yellow)$currentVersion(set_color normal)
    _log_info "Next version:   " (set_color green)$nextVersion(set_color normal)
    echo

    # Confirm release
    set -l confirm (_prompt "Create release $nextVersion? [y/N]")
    if not string match -qi "y*" $confirm
        _log_warn "Release cancelled"
        return 0
    end

    # Create tag
    _log_info "Creating tag $nextVersion..."
    git tag $nextVersion
    if test $status -ne 0
        _log_error "Failed to create tag"
        return 1
    end
    _log_success "Tag created"

    # Push commits
    _log_info "Pushing commits to origin..."
    git push origin main
    if test $status -ne 0
        _log_error "Failed to push commits"
        return 1
    end
    _log_success "Commits pushed"

    # Push tag
    _log_info "Pushing tag to origin..."
    git push origin $nextVersion
    if test $status -ne 0
        _log_error "Failed to push tag"
        return 1
    end
    _log_success "Tag pushed"

    echo
    _log_success "Release $nextVersion created successfully!"
    _log_info "GitHub Actions will now build and publish the release"
end

createRelease
