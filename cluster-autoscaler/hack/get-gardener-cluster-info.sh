function parse_flags() {
  while test $# -gt 0; do
    case "$1" in
    --shoot | -t)
      shift
      SHOOT="$1"
      ;;
    --project | -p)
      shift
      PROJECT="$1"
      ;;
    --landscape | -l)
      shift
      LANDSCAPE="$1"
      ;;
    --path | -pt)
      shift
      SAVE_PATH="$1"
      ;;
    esac
    shift
  done
}

function main(){

  parse_flags "$@"

  if [[ -z "${SHOOT}" ]]; then
    echo -e "Shoot has not been passed. Please provide Shoot either by specifying --shoot or -t argument"
    exit 1
  fi
  if [[ -z "${PROJECT}" ]]; then
    echo -e "Project has not been passed. Please provide Project either by specifying --project or -l argument"
    exit 1
  fi
  if [[ -z "${LANDSCAPE}" ]]; then
    echo -e "LANDSCAPE has not been passed. Please provide Landscape either by specifying --landscape or -p argument"
    exit 1
  fi
  if [[ -z "${SAVE_PATH}" ]]; then
    echo -e "PATH has not been passed. Please provide Path to save files either by specifying --path or -pt argument"
    exit 1
  fi

  gardenctl target --garden sap-landscape-$LANDSCAPE

  eval $(gardenctl kubectl-env bash)

  kubectl get shoot -n garden-$PROJECT $SHOOT -ojson > $SAVE_PATH/$SHOOT.json

  gardenctl target --garden sap-landscape-$LANDSCAPE --project $PROJECT --shoot $SHOOT --control-plane

  eval $(gardenctl kubectl-env bash)

  kubectl get mcc -ojson > $SAVE_PATH/$SHOOT-mcc.json

  kubectl get mcd -ojson > $SAVE_PATH/$SHOOT-mcd.json

  gardenctl target unset garden
}


main "$@"