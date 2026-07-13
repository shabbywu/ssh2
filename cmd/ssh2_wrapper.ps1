function global:showsessions {
    & ssh2 get --kind Session --template "{{ .Tag }}"
}

function global:ssh2_verify_go2s_ssh_tag {
    param(
        [Parameter(Mandatory = $true)]
        [string]$SshTag
    )

    $sessionTags = @(showsessions)
    if ($sessionTags -cnotcontains $SshTag) {
        Write-Error "Error: Session not found with tag<'$SshTag'>"
        return $false
    }
    return $true
}

function global:go2s {
    param(
        [Alias("d")]
        [switch]$Direct,

        [Parameter(Position = 0, ValueFromRemainingArguments = $true)]
        [string[]]$Arguments
    )

    $remaining = @($Arguments)
    if ($remaining.Count -gt 0 -and $remaining[0] -eq "--direct") {
        $Direct = $true
        if ($remaining.Count -gt 1) {
            $remaining = @($remaining[1..($remaining.Count - 1)])
        }
        else {
            $remaining = @()
        }
    }

    $sshTag = if ($remaining.Count -gt 0) { $remaining[0] } else { "" }
    if ([string]::IsNullOrEmpty($sshTag)) {
        showsessions
        return
    }
    if (-not (ssh2_verify_go2s_ssh_tag $sshTag)) {
        return
    }

    $ssh2Arguments = @("login")
    if ($Direct) {
        $ssh2Arguments += "--direct"
    }
    $ssh2Arguments += $sshTag
    & ssh2 @ssh2Arguments
}

Register-ArgumentCompleter -CommandName go2s -ParameterName Arguments -ScriptBlock {
    param($commandName, $parameterName, $wordToComplete, $commandAst, $fakeBoundParameters)

    $prefix = [string]$wordToComplete
    @(showsessions) | ForEach-Object {
        $sessionTag = [string]$_
        if ($sessionTag.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
            [System.Management.Automation.CompletionResult]::new(
                $sessionTag,
                $sessionTag,
                [System.Management.Automation.CompletionResultType]::ParameterValue,
                $sessionTag
            )
        }
    }
}
