source ~/.vimrc

call plug#begin(has('nvim') ? stdpath('data') . '/plugged' : '~/.vim/plugged')

Plug 'bling/vim-bufferline' " Better buffer management
Plug 'vim-airline/vim-airline' " Better statusline management
"Plug 'qwertologe/nextval.vim' " Better incrementation
"Plug 'LaTeX-Box-Team/LaTeX-Box' " LaTeX editing features
Plug 'cespare/vim-toml' " TOML syntax hilighting
Plug 'jiangmiao/auto-pairs' " Automatically pair {}, [], etc
Plug 'tpope/vim-surround' " Surround words with quotes/braces
"Plug 'rust-lang/rust.vim' " Rust syntax hilighting
"Plug 'racer-rust/vim-racer' " Rust tab completion
"let g:racer_cmd = "/Users/firechant/.cargo/bin/racer" " set racer cmd path
"let g:racer_experimental_completer = 1
"Plug 'junegunn/fzf' "multi entry selection
"Plug 'Shougo/echodoc.vim' "function signature, inline, etc
"Plug 'kshenoy/vim-signature' "show marks
"Plug 'neomake/neomake' "Linting for C++
Plug 'mechatroner/rainbow_csv' "Rainbow CSV

call plug#end()

" bufferline settings -- currently set for statusline
let g:bufferline_echo = 0
autocmd VimEnter *
        \ let &statusline='%{bufferline#refresh_status()}' 
        \ .bufferline#get_status_string()

" statusline settings
let g:airline_section_z = '(%4l,%3v)  %{airline_symbols.linenr} %L'
let g:airline#extensions#default#layout = [
            \ ['a', 'c'],
            \ ['z', 'error', 'warning']
            \ ]
let g:airline#extensions#bufferline#enabled = 1
"let g:airline#extensions#bufferline#overwrite_variables = 0
let laststatus=2 "statusline appears even before split

let g:airline#extensions#wordcount#enabled = 0
let g:airline#extensions#whitespace#enabled = 0

