package cmd

import (
	"bytes"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func newCompletionCmd() *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Generates shell completion scripts",
		Long: `To load completion run

. <(vaultpal completion bash)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(vaultpal completion bash)

# ~/.zshrc
. <(vaultpal completion zsh)
`,
		Run: nil,
	}

	zshCompletionCmd := &cobra.Command{
		Use:   "zsh",
		Short: "Completion for zsh",
		Long:  "Generate completion for zsh",
		Run: func(cmd *cobra.Command, args []string) {
			runCompletionZsh(os.Stdout)
		},
	}

	bashCompletionCmd := &cobra.Command{
		Use:   "bash",
		Short: "Completion for bash",
		Long:  "Generate completion for bash",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
	}

	completionCmd.AddCommand(zshCompletionCmd)
	completionCmd.AddCommand(bashCompletionCmd)

	return completionCmd
}

func init() {
	rootCmd.AddCommand(newCompletionCmd())
}

// Mainly copied from https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/completion/completion.go
func runCompletionZsh(out io.Writer) error {
	zshHead := "#compdef vaultpal\n"

	out.Write([]byte(zshHead))

	zshInitialization := `
__vaultpal_bash_source() {
    alias vault='env -u COMP_LINE vault'
    alias shopt=':'
    alias _expand=_bash_expand
    alias _complete=_bash_comp
    emulate -L sh
    setopt kshglob noshglob braceexpand

    source "$@"
}

__vaultpal_type() {
    # -t is not supported by zsh
    if [ "$1" == "-t" ]; then
        shift

        # fake Bash 4 to disable "complete -o nospace". Instead
        # "compopt +-o nospace" is used in the code to toggle trailing
        # spaces. We don't support that, but leave trailing spaces on
        # all the time
        if [ "$1" = "__vaultpal_compopt" ]; then
            echo builtin
            return 0
        fi
    fi
    type "$@"
}

__vaultpal_compgen() {
    local completions w
    completions=( $(compgen "$@") ) || return $?

    # filter by given word as prefix
    while [[ "$1" = -* && "$1" != -- ]]; do
        shift
        shift
    done
    if [[ "$1" == -- ]]; then
        shift
    fi
    for w in "${completions[@]}"; do
        if [[ "${w}" = "$1"* ]]; then
            echo "${w}"
        fi
    done
}

__vaultpal_compopt() {
    true # don't do anything. Not supported by bashcompinit in zsh
}

__vaultpal_ltrim_colon_completions()
{
    if [[ "$1" == *:* && "$COMP_WORDBREAKS" == *:* ]]; then
        # Remove colon-word prefix from COMPREPLY items
        local colon_word=${1%${1##*:}}
        local i=${#COMPREPLY[*]}
        while [[ $((--i)) -ge 0 ]]; do
            COMPREPLY[$i]=${COMPREPLY[$i]#"$colon_word"}
        done
    fi
}

__vaultpal_get_comp_words_by_ref() {
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[${COMP_CWORD}-1]}"
    words=("${COMP_WORDS[@]}")
    cword=("${COMP_CWORD[@]}")
}

__vaultpal_filedir() {
    local RET OLD_IFS w qw

    __vaultpal_debug "_filedir $@ cur=$cur"
    if [[ "$1" = \~* ]]; then
        # somehow does not work. Maybe, zsh does not call this at all
        eval echo "$1"
        return 0
    fi

    OLD_IFS="$IFS"
    IFS=$'\n'
    if [ "$1" = "-d" ]; then
        shift
        RET=( $(compgen -d) )
    else
        RET=( $(compgen -f) )
    fi
    IFS="$OLD_IFS"

    IFS="," __vaultpal_debug "RET=${RET[@]} len=${#RET[@]}"

    for w in ${RET[@]}; do
        if [[ ! "${w}" = "${cur}"* ]]; then
            continue
        fi
        if eval "[[ \"\${w}\" = *.$1 || -d \"\${w}\" ]]"; then
            qw="$(__vaultpal_quote "${w}")"
            if [ -d "${w}" ]; then
                COMPREPLY+=("${qw}/")
            else
                COMPREPLY+=("${qw}")
            fi
        fi
    done
}

__vaultpal_quote() {
    if [[ $1 == \'* || $1 == \"* ]]; then
        # Leave out first character
        printf %q "${1:1}"
    else
    printf %q "$1"
    fi
}

autoload -U +X bashcompinit && bashcompinit

# use word boundary patterns for BSD or GNU sed
LWORD='[[:<:]]'
RWORD='[[:>:]]'
if sed --help 2>&1 | grep -q GNU; then
    LWORD='\<'
    RWORD='\>'
fi

__vaultpal_convert_bash_to_zsh() {
    sed \
    -e 's/vaultout\[\*\]:2/vaultout\[2,\$\{#vaultout\[@\]\}\]/' \
    -e 's/declare -F/whence -w/' \
    -e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
    -e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
    -e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
    -e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
    -e "s/${LWORD}_filedir${RWORD}/__vaultpal_filedir/g" \
    -e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__vaultpal_get_comp_words_by_ref/g" \
    -e "s/${LWORD}__ltrim_colon_completions${RWORD}/__vaultpal_ltrim_colon_completions/g" \
    -e "s/${LWORD}compgen${RWORD}/__vaultpal_compgen/g" \
    -e "s/${LWORD}compopt${RWORD}/__vaultpal_compopt/g" \
    -e "s/${LWORD}declare${RWORD}/builtin declare/g" \
    -e "s/\\\$(type${RWORD}/\$(__vaultpal_type/g" \
    <<'BASH_COMPLETION_EOF'
`
	out.Write([]byte(zshInitialization))

	buf := new(bytes.Buffer)
	rootCmd.GenBashCompletion(buf)
	out.Write(buf.Bytes())

	zshTail := `
BASH_COMPLETION_EOF
}

__vaultpal_bash_source <(__vaultpal_convert_bash_to_zsh)
_complete vaultpal 2>/dev/null
`
	out.Write([]byte(zshTail))
	return nil
}
